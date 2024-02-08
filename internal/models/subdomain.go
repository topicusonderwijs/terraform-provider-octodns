package models

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"maps"
	"slices"
	"strings"
)

type Subdomain struct {
	Name        string
	keyNode     *yaml.Node
	ContentNode *yaml.Node
	Types       map[string]*Record
}

func (r *Subdomain) SetYaml(key, content *yaml.Node) {

	r.keyNode = key
	r.ContentNode = content

	r.Name = r.keyNode.Value
	r.Types = make(map[string]*Record, 0)

}

func (r *Subdomain) FindAllType() {

	for _, v := range TYPES {
		if v.IsEnabled() == true {
			_, _ = r.GetType(v.String())
		}
	}

}

func (r *Subdomain) UpdateYaml() (err error) {

	for _, v := range r.Types {

		if err = v.UpdateYaml(); err != nil {
			return
		}

	}
	return nil

}

func (r *Subdomain) validateRType(rtype string) (string, error) {

	err := fmt.Errorf("%s is not a valid record type", rtype)
	rtype = strings.ToUpper(strings.TrimSpace(rtype))

	if rt, ok := TYPES[rtype]; ok && rt.IsEnabled() {
		return rtype, nil
	}
	return "", err

}

func (r *Subdomain) CreateType(rtype string) (record *Record, err error) {
	var rtypeValidated string

	if rtypeValidated, err = r.validateRType(rtype); err != nil {
		return nil, err
	}

	if _, err = r.GetType(rtypeValidated); err != nil {
		if errors.Is(err, TypeNotFoundError) {
			// Can create Record Type

			emptyNode := &yaml.Node{}
			record = &Record{}
			record.Type = rtypeValidated
			record.RecordChild = emptyNode
			r.Types[rtypeValidated] = record

			switch r.ContentNode.Kind {
			case yaml.MappingNode:
				data := *r.ContentNode
				emptyList := &yaml.Node{Kind: yaml.SequenceNode}
				emptyList.Content = []*yaml.Node{r.ContentNode}
				err = r.ContentNode.Encode(emptyList)
				if err != nil {
					return record, fmt.Errorf("error while encoding EmptyList: %s", err.Error())
				}
				r.ContentNode.Content[0].Content = data.Content
				//r.ContentNode.Content[1].Content = emptyNode.Content

				for k := range r.Types {
					tmp, err := r.GetType(k)
					if err == nil {
						r.Types[k].RecordChild = tmp.RecordChild
					}
				}

				r.ContentNode.Content = append(r.ContentNode.Content, emptyNode)
				//err = r.ContentNode.Encode([]yaml.Node{*r.ContentNode, *emptyNode})
			case yaml.SequenceNode:
				r.ContentNode.Content = append(r.ContentNode.Content, emptyNode)
			default:
				return nil, fmt.Errorf("Dont know how to add record type to a %d node", r.ContentNode.Kind)
			}

			return record, nil
		}
	}

	return nil, TypeAlreadyExistsError

}

func (r *Subdomain) GetType(rtype string) (record *Record, err error) {

	var rtypeValidated string

	if rtypeValidated, err = r.validateRType(rtype); err != nil {
		return nil, err
	}

	if _, ok := r.Types[rtypeValidated]; ok {
		return r.Types[rtypeValidated], nil
	} else {
		yamlNode := r.findType(rtypeValidated)
		if yamlNode != nil {

			record = &Record{}
			record.RecordChild = yamlNode
			if err = yamlNode.Decode(record); err != nil {
				return nil, err
			}

			r.Types[rtypeValidated] = record
			return record, nil
		}
	}

	return nil, TypeNotFoundError

}

func (r *Subdomain) DeleteType(rtype string) (err error) {

	checkType := func(root *yaml.Node, rtype string) bool {
		for i := 0; i < len(root.Content); i += 2 {
			if root.Content[i].Value == "type" {
				if root.Content[i+1].Value == strings.ToUpper(rtype) {
					return true
				}
			}
		}
		return false
	}

	if _, ok := r.Types[rtype]; ok {
		r.Types[rtype].IsDeleted = true // Set deleted is true
		maps.DeleteFunc(r.Types, func(k string, v *Record) bool {
			if k == rtype {
				return true
			} else {
				return false
			}
		})

	}

	switch r.ContentNode.Kind {
	case yaml.MappingNode:
		if checkType(r.ContentNode, rtype) {

			r.ContentNode = &yaml.Node{}
			return nil
		}

	case yaml.SequenceNode:
		for y := 0; y < len(r.ContentNode.Content); y += 1 {
			if checkType(r.ContentNode.Content[y], rtype) {
				r.ContentNode.Content = slices.Delete(r.ContentNode.Content, y, y+1)
				return nil
			}
		}
	}

	return fmt.Errorf("type '%s' not found", rtype)

}

func (r *Subdomain) findType(rtype string) *yaml.Node {

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

	var rrecord *yaml.Node

	switch r.ContentNode.Kind {
	case yaml.MappingNode:
		rrecord = findType(r.ContentNode, rtype)
		return rrecord
	case yaml.SequenceNode:
		for y := 0; y < len(r.ContentNode.Content); y += 1 {
			if rrecord = findType(r.ContentNode.Content[y], rtype); rrecord != nil {
				return rrecord
			}
		}
	}

	return nil
}
