package models

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type Zone struct {
	name  string `yaml:"-"`
	scope string `yaml:"-"`
	doc   yaml.Node
	sha   string
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

	keyNode := yaml.Node{Kind: yaml.ScalarNode}
	keyNode.Value = subdomain
	contentNode := yaml.Node{Kind: yaml.SequenceNode}

	z.doc.Content[0].Content = append(z.doc.Content[0].Content, &keyNode, &contentNode)

	sub = Subdomain{
		Name:        subdomain,
		keyNode:     &keyNode,
		ContentNode: &contentNode,
		Types:       make(map[string]*Record),
	}

	return
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

