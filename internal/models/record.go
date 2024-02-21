package models

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
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
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()
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

func (r *Record) AddValueFromString(valueString string) error {

	var value RecordValue
	var err error

	switch r.Type {
	default:
		err = value.UnmarshalString(valueString)

	case TYPE_A.String():
		err = value.UnmarshalStringA(valueString)
	case TYPE_AAAA.String():
		err = value.UnmarshalStringAAAA(valueString)
	case TYPE_CAA.String():
		err = value.UnmarshalStringCAA(valueString)
	case TYPE_CNAME.String():
		err = value.UnmarshalStringFQDN(valueString)
	case TYPE_DNAME.String():
		err = value.UnmarshalStringFQDN(valueString)
	case TYPE_LOC.String():
		err = value.UnmarshalStringLOC(valueString)
	case TYPE_MX.String():
		err = value.UnmarshalStringMX(valueString)
	case TYPE_NAPTR.String():
		err = value.UnmarshalStringNAPTR(valueString)
	case TYPE_NS.String():
		err = value.UnmarshalStringNS(valueString)
	case TYPE_PTR.String():
		err = value.UnmarshalStringFQDN(valueString)
	case TYPE_SPF.String():
		err = value.UnmarshalStringSPF(valueString)
	case TYPE_SRV.String():
		err = value.UnmarshalStringSRV(valueString)
	case TYPE_SSHFP.String():
		err = value.UnmarshalStringSSHFP(valueString)
	case TYPE_TXT.String():
		err = value.UnmarshalString(valueString)
	case TYPE_URLFWD.String():
		err = value.UnmarshalStringURLFWD(valueString)
	}

	r.Values = append(r.Values, value)

	return err
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

func (r RecordValue) MarshalYAML() (interface{}, error) {

	if r.StringValue != nil {
		return *r.StringValue, nil
	} else {
		return r.baseRecordValue, nil
	}
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

	if r.LatDegrees == nil || r.LongDegrees == nil || r.LatDirection == nil || r.LongDirection == nil || r.Altitude == nil {
		return ""
	}

	fnLatLong := func(degrees, minutes *int, seconds *float64, direction *string) string {
		if degrees == nil || direction == nil {
			return ""
		}

		ret := []string{fmt.Sprintf("%d", *degrees)}
		if minutes != nil {
			ret = append(ret, fmt.Sprintf("%d", *minutes))
			if seconds != nil {
				ret = append(ret, fmt.Sprintf("%0.2f", *seconds))
			}
		}
		ret = append(ret, *direction)

		return strings.Join(ret, " ")

	}

	//Lat: d1 [m1 [s1]] {"N"|"S"}
	//Long: d2 [m2 [s2]] {"E"|"W"}
	//Altitude: alt["m"]
	//optional: [siz["m"] [hp["m"] [vp["m"]]]
	//]

	optional := ""
	if r.Size != nil {
		if r.PrecisionHorz != nil {
			if r.PrecisionVert != nil {
				optional = fmt.Sprintf("%d %d %d", *r.Size, *r.PrecisionHorz, *r.PrecisionVert)
			} else {
				optional = fmt.Sprintf("%d %d", *r.Size, *r.PrecisionHorz)
			}
		} else {
			optional = fmt.Sprintf("%d", *r.Size)
		}
	}

	return fmt.Sprintf(
		"%s %s %0.2f %s",
		fnLatLong(r.LatDegrees, r.LatMinutes, r.LatSeconds, r.LatDirection),
		fnLatLong(r.LongDegrees, r.LongMinutes, r.LongSeconds, r.LongDirection),
		*r.Altitude,
		optional,
	)

}

func (r *RecordValue) validateIP(ip string) error {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ValidateNotAnIP
	}
	return nil
}

func (r *RecordValue) validateIPV4(ip string) error {
	if err := r.validateIP(ip); err != nil {
		return ValidateNotAnIPV4
	}
	if !strings.Contains(ip, ".") || strings.Contains(ip, ":") {
		return ValidateNotAnIPV4
	}

	return nil
}
func (r *RecordValue) validateIPV6(ip string) error {
	if err := r.validateIP(ip); err != nil {
		return ValidateNotAnIPV6
	}
	if !strings.Contains(ip, ":") || strings.Contains(ip, ".") {
		return ValidateNotAnIPV6
	}

	return nil
}
func (r *RecordValue) validateFQDN(value string, requireDot bool) error {
	if err := r.validateIP(value); err == nil {
		return ValidateIPNotAllowed
	}

	if !strings.HasSuffix(value, ".") && requireDot {
		return ValidateFQDNRequiredTrailingDot
	} else if strings.HasSuffix(value, ".") && !requireDot {
		return ValidateFQDNForbidTrailingDot
	}

	reg, err := regexp.Compile(`([\pL\pN\pS\-\_\.])+(\.?([\pL\pN]|xn\-\-[\pL\pN-]+)+\.?)`)
	if err != nil {
		return err
	}
	if !reg.MatchString(value) {
		return ValidateNotAFQDN
	}

	return nil
}

func (r *RecordValue) UnmarshalString(value string) error {

	if strings.Count(value, ";") != strings.Count(value, "\\;") {
		return fmt.Errorf("value contains unescaped ';'")
	}

	r.baseRecordValue.StringValue = &value
	return nil
}

func (r *RecordValue) UnmarshalStringA(value string) error {
	if err := r.validateIPV4(value); err != nil {
		return err
	}
	return r.UnmarshalString(value)
}

func (r *RecordValue) UnmarshalStringAAAA(value string) error {
	if err := r.validateIPV6(value); err != nil {
		return err
	}

	return r.UnmarshalString(value)
}

func (r *RecordValue) UnmarshalStringFQDN(value string) error {

	if err := r.validateFQDN(value, true); err != nil {
		return err
	}

	return r.UnmarshalString(value)
	//if r.StringValue != nil {
	//	return *r.StringValue
	//} else {
	//	return ""
	//}
}

func (r *RecordValue) UnmarshalStringMX(value string) error {

	parts, err := regexToMap(value, `^(?P<preference>\d+) (?P<exchange>.+)$`)
	if err != nil {
		return err
	}

	pref, err := strconv.Atoi(parts["preference"])
	if err != nil {
		return fmt.Errorf("preference should be a int value")
	}

	err = r.validateFQDN(parts["exchange"], true)
	if err != nil {
		return err
	}

	r.baseRecordValue.Preference = RefInt(pref)
	r.baseRecordValue.Exchange = RefString(parts["exchange"])

	return nil
}

func (r *RecordValue) UnmarshalStringSRV(value string) error {

	parts, err := regexToMap(value, `^(?P<priority>\d+) (?P<weight>\d+) (?P<port>\d+) (?P<target>.+)$`)
	if err != nil {
		return err
	}

	intPrio, _ := strconv.Atoi(parts["priority"])
	intWeight, _ := strconv.Atoi(parts["weight"])
	intPort, _ := strconv.Atoi(parts["port"])

	r.Priority = RefInt(intPrio)
	r.Weight = RefInt(intWeight)
	r.Port = RefInt(intPort)

	errFQDN := r.validateFQDN(parts["target"], true)
	errIP := r.validateIP(parts["target"])
	if errFQDN != nil && errIP != nil {

		return fmt.Errorf("target must be a FQDN or IP")
	}

	r.Target = RefString(parts["target"])

	return nil

}

func (r *RecordValue) UnmarshalStringCAA(value string) error {

	parts, err := regexToMap(value, `^(?P<flags>\d+) (?P<tag>issue|issuewild|iodef) (?P<target>(?P<host>.+)(?P<policy>|; policy=(dv|ev|cv)))$`)
	if err != nil {
		return err
	}

	intFlags, err := strconv.Atoi(parts["flags"])
	if err != nil || intFlags < 0 || intFlags > 128 {
		return fmt.Errorf("flag part should be an int between 0 and 128")
	}

	if parts["tag"] != "iodef" && r.validateFQDN(parts["host"], false) != nil {
		return fmt.Errorf("issuer should be a valid hostname")
	}
	if parts["tag"] == "iodef" && !strings.Contains(parts["host"], "mailto:") {
		return fmt.Errorf("issuer should be a valid mailto link")
	}

	r.Tag = RefString(parts["tag"])
	r.Flags = RefString(parts["flags"])
	r.Value = RefString(parts["target"])

	return nil

}

func (r *RecordValue) UnmarshalStringURLFWD(value string) error {

	parts, err := regexToMap(value, `^(?P<code>0|301|302) (?P<masking>0|1|2) (?P<path>[^ ]+) (?P<query>0|1) (?P<target>.+)$`)
	if err != nil {
		return err
	}

	if strings.HasSuffix(parts["path"], "/") && parts["path"] != "/" {
		return fmt.Errorf("Path must not end with a slash (/)")
	}

	//@todo: Validate values
	r.Code = RefStringAsInt(parts["code"])
	r.Masking = RefStringAsInt(parts["masking"])
	r.Path = RefString(parts["path"])
	r.Query = RefStringAsInt(parts["query"])
	r.Target = RefString(parts["target"])

	return nil

}
func (r *RecordValue) UnmarshalStringSSHFP(value string) error {

	parts, err := regexToMap(value, `^(?P<algorithm>\d+) (?P<fingerprinttype>\d+) (?P<fingerprint>.+)$`)
	if err != nil {
		return err
	}

	//@todo: Validate values
	r.Algorithm = RefStringAsInt(parts["algorithm"])
	r.FingerprintType = RefStringAsInt(parts["fingerprinttype"])
	r.Fingerprint = RefString(parts["fingerprint"])

	return nil

}

func (r *RecordValue) UnmarshalStringNAPTR(value string) error {

	parts, err := regexToMap(value, `^(?P<order>\d+) (?P<preference>\d+) \"(?P<flags>.+)\" \"(?P<service>.+)\" \"(?P<regexp>.+)\" (?P<replacement>.+)$`)
	if err != nil {
		return err
	}

	//@todo: Validate values
	r.Order = RefStringAsInt(parts["order"])
	r.Preference = RefStringAsInt(parts["preference"])
	r.Flags = RefString(parts["flags"])
	r.Service = RefString(parts["service"])
	r.Regexp = RefString(parts["regexp"])
	r.Replacement = RefString(parts["replacement"])

	return nil
}

func (r *RecordValue) UnmarshalStringLOC(value string) error {

	/*
		return fmt.Sprintf(
			"%s %s %0.2f %s",
			fnLatLong(r.LatDegrees, r.LatMinutes, r.LatSeconds, r.LatDirection),
			fnLatLong(r.LongDegrees, r.LongMinutes, r.LongSeconds, r.LongDirection),
			*r.Altitude,
			optional,
		)
	*/
	/*
		values := []string{
			"31 58 52.10 S 115 49 11.70 E 20.00 10 10 2",
			"31 58 S 115 49 E 20.00 10 10 2",
			"31 S 115 E 20.00 10 10 2",
			"31 S 115 E 20.00 10 10",
			"31 S 115 E 20.00 10",
			"31 S 115 E 20.00",
		}

		for _, value = range values {
	*/

	//(?:\d+(?:\.\d*)?|\.\d+)
	parts, err := regexToMap(
		value,
		`^(?P<latdeg>\d+) (|(?P<latm>\d+) (|(?P<lats>\d+(|.\d+)) ))(?P<latdir>N|S) `+
			`(?P<longdeg>\d+) (|(?P<longm>\d+) (|(?P<longs>\d+(|.\d+)) ))(?P<longdir>E|W) `+
			`(?P<alt>\d+(|.\d+))`+
			`(| (?P<size>\d+)(| (?P<prh>\d+)(| (?P<prv>\d+))))$`,
	)
	if err != nil {
		return err
	}

	//@todo: Validate values
	r.LatDegrees = RefStringAsInt(parts["latdeg"])
	r.LatMinutes = RefStringAsInt(parts["latm"])
	r.LatSeconds = RefStringAsFloat(parts["lats"])
	r.LatDirection = RefString(parts["latdir"])

	r.LongDegrees = RefStringAsInt(parts["longdeg"])
	r.LongMinutes = RefStringAsInt(parts["longm"])
	r.LongSeconds = RefStringAsFloat(parts["longs"])
	r.LongDirection = RefString(parts["longdir"])

	r.Altitude = RefStringAsFloat(parts["alt"])
	r.Size = RefStringAsInt(parts["size"])
	r.PrecisionHorz = RefStringAsInt(parts["prh"])
	r.PrecisionVert = RefStringAsInt(parts["prv"])

	return nil

}
func (r *RecordValue) UnmarshalStringNS(value string) error {

	errIP := r.validateIP(value)
	errHN := r.validateFQDN(value, true)

	if errIP != nil && errHN != nil {
		return fmt.Errorf("value should be a valid IP of FQDN")
	}

	return r.UnmarshalString(value)
}
func (r *RecordValue) UnmarshalStringPTR(value string) error {
	return r.UnmarshalString(value)

}
func (r *RecordValue) UnmarshalStringSPF(value string) error {

	if !strings.HasPrefix(value, "v=spf1") {
		return fmt.Errorf("value should start with v=spf1")
	}

	reg, err := regexp.Compile(`((-|~|\+)all)`)
	if err != nil {
		return err
	}

	if !reg.MatchString(value) {
		return fmt.Errorf("value should contain a part for ALL")
	}

	return r.UnmarshalString(value)
}
