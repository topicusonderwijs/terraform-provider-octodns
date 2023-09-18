package models

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type Zone struct {
	doc yaml.Node
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
		//fmt.Println("Problem key found")
		content = bytes.Replace(content, []byte("? ''\n  :"), []byte("'':\n   "), 1)
	}
	err := yaml.Unmarshal(content, &z.doc)
	if err != nil {
		return err
	}

	return nil

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
			Name:         "",
			Type:         "",
			Values:       nil,
			TTL:          0,
			Terraform:    Terraform{},
		},
	}

	_ = recordChild
	_ = recordNode
	_ = recordParent

	err = recordChild.Decode(record)
	if err != nil {
		return
	}

	return
}

func (z *Zone) FindRecord(subdomain string) (record Subdomain, err error) {

	if subdomain == "@" {
		subdomain = ""
	}

	if z.doc.Kind != yaml.DocumentNode {
		err = fmt.Errorf("z.doc is not a document node")
		return
	}

	if len(z.doc.Content[0].Content) == 0 {
		err = fmt.Errorf(z.doc.Content[0].Value)
		return
	}

	for i := 0; i < len(z.doc.Content[0].Content); i += 2 {
		/*
			switch z.doc.Content[0].Kind {
			case yaml.DocumentNode:
				fmt.Println("DocumentNode")
			case yaml.MappingNode:
				fmt.Println("MappingNode")
			case yaml.SequenceNode:
				fmt.Println("SequenceNode")
			}
		*/

		//		log.Print("Found: ", z.doc.Content[0].Content[i].Value, " ", subdomain, " ", z.doc.Content[0].Content[i].Value == subdomain)

		if z.doc.Content[0].Content[i].Value == subdomain {

			record.SetYaml(z.doc.Content[0].Content[i], z.doc.Content[0].Content[i+1])
			return record, nil
		}
	}

	return record, fmt.Errorf("subdomain not found")
}

func (z *Zone) FindRecordByType(subdomain string, rtype string) (rrecord *yaml.Node, rcontent *yaml.Node, rparent *yaml.Node, err error) {

	if z.doc.Kind != yaml.DocumentNode {
		err = fmt.Errorf("z.doc is not a document node")
		return
	}

	for i := 0; i < len(z.doc.Content[0].Content); i += 2 {
		/*
			switch z.doc.Content[0].Kind {
			case yaml.DocumentNode:
				fmt.Println("DocumentNode")
			case yaml.MappingNode:
				fmt.Println("MappingNode")
			case yaml.SequenceNode:
				fmt.Println("SequenceNode")
			}
		*/

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

			_ = rparent
			_ = rcontent

			//fmt.Println("Found ", subdomain)

			switch rcontent.Kind {
			case yaml.MappingNode:
				//fmt.Println("Map")
				rrecord = findType(rcontent, rtype)
				return
			case yaml.SequenceNode:
				//fmt.Println("Seq")
				for y := 0; y < len(rcontent.Content); y += 1 {
					fmt.Println("Y", y)
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

type OldZone struct {
	Name string `yaml:"-"`

	Records []Record
}

func (z *OldZone) UnmarshalYAML(
	value *yaml.Node,
) error {
	//fmt.Println("Blaats")

	var items map[string]yaml.Node
	if err := value.Decode(&items); err == nil {
		for k, v := range items {
			if k == "" {
				k = "@"
			}

			var slice []yaml.Node
			var object yaml.Node
			if err := v.Decode(&slice); err == nil {
				// Node is Slice
			} else if err := v.Decode(&object); err == nil {
				// Node is Single object
				slice = []yaml.Node{object}
			} else {
				return err
			}

			records, err := decodeRecords(k, slice)
			if err != nil {
				return err
			}
			_ = records
			z.Records = append(z.Records, records...)
			//fmt.Println("Subdomain found:", len(records))

		}
		return nil
	} else {
		return err
	}
	return nil
}

func (z OldZone) MarshalYAML() (interface{}, error) {

	out := make(map[string][]Record, 0)

	for _, record := range z.Records {

		if _, ok := out[record.Name]; !ok {
			out[record.Name] = []Record{record}
		} else {
			out[record.Name] = append(out[record.Name], record)
		}

	}

	return out, nil

}

func decodeRecords(subdomain string, nodes []yaml.Node) ([]Record, error) {

	records := []Record{}

	for i := range nodes {
		record, err := decodeRecord(subdomain, nodes[i])
		if err != nil {
			return []Record{}, err
		}
		records = append(records, record)
	}

	return records, nil

}

func decodeRecord(subdomain string, node yaml.Node) (Record, error) {

	record := Record{}
	record.Name = subdomain

	err := node.Decode(&record)

	return record, err

}

/*
func (z *Record) UnmarshalYAML(
	value *yaml.Node,
) error {
	fmt.Println("2Blaat")
	var name string
	if err := value.Decode(&name); err == nil {
		fmt.Println("2Blaat", name)
		return nil
	} else {
		fmt.Println("2Error: ", err.Error())
	}
	return nil
}
*/
/*
func (z *OldZone) UnmarshalYAML(

	unmarshal func(interface{}) error,

	) error {
		var name string
		if err := unmarshal(&name); err == nil {
			fmt.Println("Blaat", name)
			return nil
		} else {
			fmt.Println("Error: ", err.Error())
		}
		return nil
	}
*/
