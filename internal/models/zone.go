package models

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Zone struct {
	name  string `yaml:"-"`
	scope string `yaml:"-"`
	doc   yaml.Node
	sha   string

	// CompactSequenceIndent writes sequences without indentation relative to
	// their parent mapping key (`key:\n- item`) instead of `key:\n  - item`.
	CompactSequenceIndent bool
}

func (z *Zone) ReadYamlFile(filename string) error {

	//@todo: Check if file path is sane/existing/readable/etc

	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return z.ReadYaml(fileContent)
}

func (z *Zone) ReadYaml(content []byte) error {

	if bytes.Contains(content, []byte("? ''\n  :")) {
		content = bytes.Replace(content, []byte("? ''\n  :"), []byte("'':\n   "), 1)
	}
	err := yaml.Unmarshal(content, &z.doc)
	if err != nil {
		return err
	}

	return nil

}

func (z *Zone) CreateSubdomain(subdomain string) (sub Subdomain, err error) {

	var existingSub Subdomain
	existingSub, err = z.FindSubdomain(subdomain)
	if err == nil {
		return existingSub, ErrSubdomainAlreadyExists
	} else {
		if errors.Is(err, ErrSubdomainNotFound) {
			err = nil // We can ignore this error
		} else {
			return // return original error
		}
	}

	keyNode := yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str"}
	keyNode.Value = subdomain
	// Single quote keys that would not be read back as a string, like 9.50 or
	// 10 (common in reverse zones), so other tools do not parse them as numbers
	if !isPlainYamlString(subdomain) {
		keyNode.Style = yaml.SingleQuotedStyle
	}
	contentNode := yaml.Node{Kind: yaml.SequenceNode}

	// An empty zone is usually written as {}, switch to block style so new
	// records are not written as flow style yaml
	if len(z.doc.Content[0].Content) == 0 {
		z.doc.Content[0].Style &^= yaml.FlowStyle
	}

	z.doc.Content[0].Content = append(z.doc.Content[0].Content, &keyNode, &contentNode)

	sub = Subdomain{
		Name:        subdomain,
		keyNode:     &keyNode,
		ContentNode: &contentNode,
		Types:       make(map[string]*Record),
	}

	return
}

// isPlainYamlString reports whether value is read back as a string when
// written as a plain (unquoted) yaml scalar.
func isPlainYamlString(value string) bool {
	var v any
	if err := yaml.Unmarshal([]byte(value), &v); err != nil {
		return false
	}
	_, ok := v.(string)
	return ok
}

func (z *Zone) GetRecord(subdomain string, rtype string) (record *Record, err error) {

	recordChild, recordNode, recordParent, err := z.FindRecordByType(subdomain, rtype)
	if err != nil {
		return nil, err
	}

	record = &Record{
		BaseRecord: BaseRecord{
			RecordChild:  recordChild,
			RecordNode:   recordNode,
			RecordParent: recordParent,
			Name:         subdomain,
			Type:         "",
			Values:       nil,
			TTL:          0,
			Terraform:    Terraform{},
		},
	}

	err = recordChild.Decode(record)
	if err != nil {
		return
	}

	return
}

func (z *Zone) DeleteSubdomain(subdomain string) (err error) {

	if subdomain == "@" {
		subdomain = ""
	}

	if z.doc.Kind != yaml.DocumentNode {
		err = fmt.Errorf("zone.doc is not a document node")
		return
	}

	if len(z.doc.Content[0].Content) == 0 {
		err = fmt.Errorf("error: %s", z.doc.Content[0].Value)
		return
	}

	for i := 0; i < len(z.doc.Content[0].Content); i += 2 {
		if z.doc.Content[0].Content[i].Value == subdomain {
			z.doc.Content[0].Content = slices.Delete(z.doc.Content[0].Content, i, i+2)
			return nil
		}
	}
	return nil
}

func (z *Zone) FindSubdomain(subdomain string) (record Subdomain, err error) {

	if subdomain == "@" {
		subdomain = ""
	}

	if z.doc.Kind != yaml.DocumentNode {
		err = fmt.Errorf("zone.doc is not a document node")
		return
	}

	if len(z.doc.Content[0].Content) == 0 {
		// An empty zone ({}) simply does not contain the subdomain
		if z.doc.Content[0].Kind == yaml.MappingNode {
			err = ErrSubdomainNotFound
			return
		}
		err = fmt.Errorf("error: %s", z.doc.Content[0].Value)
		return
	}

	for i := 0; i < len(z.doc.Content[0].Content); i += 2 {
		if z.doc.Content[0].Content[i].Value == subdomain {

			record.SetYaml(z.doc.Content[0].Content[i], z.doc.Content[0].Content[i+1])
			return record, nil
		}
	}

	return record, ErrSubdomainNotFound
}

func (z *Zone) FindRecordByType(subdomain string, rtype string) (rrecord *yaml.Node, rcontent *yaml.Node, rparent *yaml.Node, err error) {

	if z.doc.Kind != yaml.DocumentNode {
		err = fmt.Errorf("zone.doc is not a document node")
		return
	}

	for i := 0; i < len(z.doc.Content[0].Content); i += 2 {

		findType := func(root *yaml.Node, rtype string) *yaml.Node {
			for i := 0; i < len(root.Content); i += 2 {
				if root.Content[i].Value == "type" {
					if root.Content[i+1].Value == strings.ToUpper(rtype) {
						return root

					}
				}
			}

			return nil
		}

		if z.doc.Content[0].Content[i].Value == subdomain {

			xRecord := Subdomain{}
			xRecord.SetYaml(z.doc.Content[0].Content[i], z.doc.Content[0].Content[i+1])

			rparent = z.doc.Content[0].Content[i]
			rcontent = z.doc.Content[0].Content[i+1]

			switch rcontent.Kind {
			case yaml.MappingNode:
				rrecord = findType(rcontent, rtype)
				return
			case yaml.SequenceNode:
				for y := 0; y < len(rcontent.Content); y += 1 {
					if rrecord = findType(rcontent.Content[y], rtype); rrecord != nil {
						return
					}
				}
			}

			return
		}

	}

	return nil, nil, nil, fmt.Errorf("subdomain not found")

}

func (z Zone) WriteYaml() ([]byte, error) {

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if z.CompactSequenceIndent {
		encoder.CompactSeqIndent()
	}
	err := encoder.Encode(z.doc.Content[0])
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (z Zone) WriteYamlToFile(filename string) error {

	data, err := z.WriteYaml()
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, data, 0666)
	if err != nil {
		return err
	}
	return nil
}
