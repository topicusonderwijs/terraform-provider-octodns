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
		case TYPE_CAA.String():
			ret = append(ret, v.StringCAA())
		case TYPE_URLFWD.String():
			ret = append(ret, v.StringURLFWD())
		case TYPE_SSHFP.String():
			ret = append(ret, v.StringSSHFP())
		case TYPE_NAPTR.String():
			ret = append(ret, v.StringNAPTR())
		case TYPE_LOC.String():
			ret = append(ret, v.StringLOC())
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
		out.Values = node
	} else {
		err = node.Encode(r.Values[0])
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
	Altitude      *float64 `yaml:",omitempty"`
	LatDegrees    *int     `yaml:"lat_degrees,omitempty"`
	LatDirection  *string  `yaml:"lat_direction,omitempty"`
	LatMinutes    *int     `yaml:"lat_minutes,omitempty"`
	LatSeconds    *float64 `yaml:"lat_seconds,omitempty"`
	LongDegrees   *int     `yaml:"long_degrees,omitempty"`
	LongDirection *string  `yaml:"long_direction,omitempty"`
	LongMinutes   *int     `yaml:"long_minutes,omitempty"`
	LongSeconds   *float64 `yaml:"long_seconds,omitempty"`
	PrecisionHorz *int     `yaml:"precision_horz,omitempty"`
	PrecisionVert *int     `yaml:"precision_vert,omitempty"`
	Size          *int     `yaml:",omitempty"`

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

func (r RecordValue) StringCAA() string {

	if r.Flags != nil && r.Tag != nil && r.Value != nil {
		return fmt.Sprintf("%s %s %s", *r.Flags, *r.Tag, *r.Value)
	} else {
		return ""
	}
}

func (r RecordValue) StringURLFWD() string {

	if r.Code != nil && r.Masking != nil && r.Path != nil && r.Query != nil && r.Target != nil {
		return fmt.Sprintf(
			"%d %d %s %d %s",
			*r.Code, *r.Masking, *r.Path, *r.Query, *r.Target,
		)
	} else {
		return ""
	}
}
func (r RecordValue) StringSSHFP() string {
	if r.Algorithm != nil && r.FingerprintType != nil && r.Fingerprint != nil {
		return fmt.Sprintf(
			"%d %d %s",
			*r.Algorithm, *r.FingerprintType, *r.Fingerprint,
		)
	} else {
		return ""
	}
}

func (r RecordValue) StringNAPTR() string {

	if r.Order != nil && r.Preference != nil && r.Flags != nil && r.Service != nil && r.Regexp != nil && r.Replacement != nil {
		return fmt.Sprintf(
			"%d %d %q %q %q %s",
			*r.Order, *r.Preference, *r.Flags, *r.Service, *r.Regexp, *r.Replacement,
		)
	} else {
		return ""
	}
}

func (r RecordValue) StringLOC() string {

	return fmt.Sprintf(
		"%d %d %0.2f %s %d %d %0.2f %s %0.2f %d %d %d",
		*r.LatDegrees,
		*r.LatMinutes,
		*r.LatSeconds,
		*r.LatDirection,
		*r.LongDegrees,
		*r.LongMinutes,
		*r.LongSeconds,
		*r.LongDirection,
		*r.Altitude,
		*r.Size,
		*r.PrecisionHorz,
		*r.PrecisionVert,
	)

}

func (r RecordValue) MarshalYAML() (interface{}, error) {

	if r.StringValue != nil {
		return *r.StringValue, nil
	} else {
		return r.baseRecordValue, nil
	}
}
