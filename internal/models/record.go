package models

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type Terraform struct {
	Hash string `yaml:",omitempty"`
}

type OctodnsRecordConfig struct {
	Cloudflare *OctodnsCloudflare `yaml:",omitempty"`
	AzureDNS   *OctodnsAzureDNS   `yaml:",omitempty"`
}

type OctodnsCloudflare struct {
	Proxied bool `yaml:",omitempty"`
	AutoTTL bool `yaml:",omitempty"`
}

type OctodnsAzureDNS struct {
	Healthcheck OctodnsAzureDNSHealthcheck `yaml:",omitempty"`
}

type OctodnsAzureDNSHealthcheck struct {
	Interval    int `yaml:",omitempty"`
	Timeout     int `yaml:",omitempty"`
	NumFailures int `yaml:",omitempty"`
}

type BaseRecord struct {
	RecordChild  *yaml.Node `yaml:"-"`
	RecordNode   *yaml.Node `yaml:"-"`
	RecordParent *yaml.Node `yaml:"-"`
	IsDeleted    bool       `yaml:"-"`
	Name         string
	Type         string
	Values       []RecordValue `yaml:"values" line_comment:"Enable or disable."`
	TTL          int
	Terraform    Terraform
	Octodns      OctodnsRecordConfig
}

type Record struct {
	BaseRecord
	Values []RecordValue `yaml:"values" line_comment:"Enable or disable."`
}

type recordYamlUnmarshal struct {
	Name      string              `yaml:",omitempty"`
	Type      string              `yaml:",omitempty"`
	Value     yaml.Node           `yaml:",omitempty"`
	Values    yaml.Node           `yaml:",omitempty"`
	TTL       int                 `yaml:",omitempty"`
	Terraform Terraform           `yaml:",omitempty"`
	Octodns   OctodnsRecordConfig `yaml:",omitempty"`
}

func (r *Record) UpdateYaml() error {
	if r.IsDeleted {
		return nil
	}
	return r.RecordChild.Encode(r)
}

func (r *Record) ValuesAsString() []string {
	var ret []string
	for _, v := range r.Values {
		switch r.Type {
		default:
			ret = append(ret, v.String())
		case TYPE_MX.String():
			ret = append(ret, v.StringMX())
		case TYPE_SRV.String():
			ret = append(ret, v.StringSRV())

		}

	}

	return ret
}

func (r *Record) ClearValues() {
	r.Values = []RecordValue{}
}

func (r *Record) AddValueFromString(valueString string) {

	var value RecordValue

	switch r.Type {
	default:
		value = RecordValue{baseRecordValue{StringValue: &valueString}}
		/*
			case TYPE_MX.String():
				value = append(ret, v.StringMX())
			case TYPE_SRV.String():
				value = append(ret, v.StringSRV())
		*/
	}

	r.Values = append(r.Values, value)

}

func (r *Record) AddType(record Record) error {
	noreRecord := yaml.Node{}

	if err := noreRecord.Encode(record); err != nil {
		return err
	}

	if r.RecordNode.Kind == yaml.SequenceNode {
		r.RecordNode.Content = append(r.RecordNode.Content, &noreRecord)
	} else {
		return fmt.Errorf("Can not add new type to a single typed record")
	}

	return nil
}

func (r *Record) UnmarshalYAML(
	value *yaml.Node,
) error {

	raw := recordYamlUnmarshal{}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	r.TTL = raw.TTL
	r.Terraform = raw.Terraform
	r.Type = raw.Type
	r.Octodns = raw.Octodns

	if !raw.Value.IsZero() {
		rvValue := RecordValue{}
		err := rvValue.UnmarshalYAML(&raw.Value)
		if err != nil {
			return err
		}
		r.Values = append(r.Values, rvValue)
	} else if !raw.Values.IsZero() {
		rvValues := make([]RecordValue, 0)
		err := raw.Values.Decode(&rvValues)
		if err != nil {
			return err
		}
		r.Values = append(r.Values, rvValues...)
	}

	return nil
}

func (r Record) MarshalYAML() (interface{}, error) {

	out := recordYamlUnmarshal{
		Type:      r.Type,
		TTL:       r.TTL,
		Terraform: r.Terraform,
		Octodns:   r.Octodns,
	}
	node := yaml.Node{}
	var err error
	if len(r.Values) == 0 {
		return nil, fmt.Errorf("0 Values encountered: %s, %s ", r.Name, r.Type)
	}
	if len(r.Values) > 1 {
		err = node.Encode(r.Values)
		node.LineComment = "Blaat"
		out.Values = node
	} else {
		err = node.Encode(r.Values[0])
		node.LineComment = "Blaat"
		out.Value = node
	}
	if err != nil {
		return nil, err
	}

	return out, nil

}

type baseRecordValue struct {
	StringValue *string `yaml:",omitempty"`
	Comment     *string `yaml:"-"`

	// SSHFP

	Algorithm       *int    `yaml:",omitempty"`
	Fingerprint     *string `yaml:",omitempty"`
	FingerprintType *int    `yaml:"fingerprint_type,omitempty"`

	Flags *string `yaml:",omitempty"`
	Tag   *string `yaml:",omitempty"`
	//Value string `yaml:",omitempty"`

	// SRV
	Port     *int    `yaml:",omitempty"`
	Priority *int    `yaml:",omitempty"`
	Target   *string `yaml:",omitempty"`
	Weight   *int    `yaml:",omitempty"`

	// MX
	Exchange   *string `yaml:",omitempty"`
	Preference *int    `yaml:",omitempty"`
	Value      *string `yaml:",omitempty"`
	//Priority   int    `yaml:",omitempty"`

	// LOC
	Altitude       *int     `yaml:",omitempty"`
	Lat_degrees    *int     `yaml:",omitempty"`
	Lat_direction  *string  `yaml:",omitempty"`
	Lat_minutes    *int     `yaml:",omitempty"`
	Lat_seconds    *float64 `yaml:",omitempty"`
	Long_degrees   *int     `yaml:",omitempty"`
	Long_direction *string  `yaml:",omitempty"`
	Long_minutes   *int     `yaml:",omitempty"`
	Long_seconds   *float64 `yaml:",omitempty"`
	Precision_horz *int     `yaml:",omitempty"`
	Precision_vert *int     `yaml:",omitempty"`
	Size           *int     `yaml:",omitempty"`

	// NAPTR
	//Flags       string `yaml:",omitempty"`
	//Preference  int    `yaml:",omitempty"`
	Order       *int    `yaml:",omitempty"`
	Regexp      *string `yaml:",omitempty"`
	Replacement *string `yaml:",omitempty"`
	Service     *string `yaml:",omitempty"`

	// URLFWD

	Code    *int    `yaml:",omitempty"`
	Masking *int    `yaml:",omitempty"`
	Path    *string `yaml:",omitempty"`
	Query   *int    `yaml:",omitempty"`
	//Target  string `yaml:",omitempty"`
}

type RecordValue struct {
	baseRecordValue
}

func (r *RecordValue) UnmarshalYAML(
	value *yaml.Node,
) error {

	if len(value.Content) == 0 {
		if err := value.Decode(&r.StringValue); err != nil {
			return err
		}
	} else {

		al := baseRecordValue{}

		if err := value.Decode(&al); err != nil {
			return err
		}

		r.baseRecordValue = al

	}

	return nil

}

func (r RecordValue) String() string {
	if r.StringValue != nil {
		return *r.StringValue
	} else {
		return ""
	}
}

func (r RecordValue) StringMX() string {

	if r.Priority != nil && r.Value != nil {
		return fmt.Sprintf("%d %s", *r.Priority, *r.Value)
	} else if r.Preference != nil && r.Exchange != nil {
		return fmt.Sprintf("%d %s", *r.Preference, *r.Exchange)
	}

	return ""
}

func (r RecordValue) StringSRV() string {

	if r.Priority == nil || r.Weight == nil || r.Port == nil || r.Target == nil {
		return ""
	} else {
		return fmt.Sprintf("%d %d %d %s", *r.Priority, *r.Weight, *r.Port, *r.Target)
	}

}

func (r RecordValue) MarshalYAML() (interface{}, error) {

	if r.StringValue != nil {
		return *r.StringValue, nil
	} else {
		return r.baseRecordValue, nil
	}
}
