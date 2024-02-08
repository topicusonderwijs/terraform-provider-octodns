package models

import "strings"

var (
	TYPE_A      RType = RType{value: "A", enabled: true}
	TYPE_AAAA   RType = RType{value: "AAAA", enabled: true}
	TYPE_CAA    RType = RType{value: "CAA", enabled: true}
	TYPE_CNAME  RType = RType{value: "CNAME", enabled: true}
	TYPE_DNAME  RType = RType{value: "DNAME", enabled: true}
	TYPE_LOC    RType = RType{value: "LOC", enabled: true}
	TYPE_MX     RType = RType{value: "MX", enabled: true}
	TYPE_NAPTR  RType = RType{value: "NAPTR", enabled: true}
	TYPE_NS     RType = RType{value: "NS", enabled: true}
	TYPE_PTR    RType = RType{value: "PTR", enabled: true}
	TYPE_SPF    RType = RType{value: "SPF", enabled: true}
	TYPE_SRV    RType = RType{value: "SRV", enabled: true}
	TYPE_SSHFP  RType = RType{value: "SSHFP", enabled: true}
	TYPE_TXT    RType = RType{value: "TXT", enabled: true}
	TYPE_URLFWD RType = RType{value: "URLFWD", enabled: true}

	TYPES = map[string]RType{
		TYPE_A.String():      TYPE_A,
		TYPE_AAAA.String():   TYPE_AAAA,
		TYPE_CAA.String():    TYPE_CAA,
		TYPE_CNAME.String():  TYPE_CNAME,
		TYPE_DNAME.String():  TYPE_DNAME,
		TYPE_LOC.String():    TYPE_LOC,
		TYPE_MX.String():     TYPE_MX,
		TYPE_NAPTR.String():  TYPE_NAPTR,
		TYPE_NS.String():     TYPE_NS,
		TYPE_PTR.String():    TYPE_PTR,
		TYPE_SPF.String():    TYPE_SPF,
		TYPE_SRV.String():    TYPE_SRV,
		TYPE_SSHFP.String():  TYPE_SSHFP,
		TYPE_TXT.String():    TYPE_TXT,
		TYPE_URLFWD.String(): TYPE_URLFWD,
	}
)

type RType struct {
	value   string
	enabled bool
}

func (r RType) IsEnabled() bool {
	return r.enabled
}

func (r RType) String() string {
	return r.value
}

func (r RType) LowerString() string {
	return strings.ToLower(r.value)
}
