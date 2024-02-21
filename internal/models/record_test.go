package models

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"os"
	"strings"
	"testing"
)

type TestCheck struct {
	Type        RType
	ReadTested  bool
	WriteTested bool
}

func (c *TestCheck) Read() {
	c.ReadTested = true
}

func (c *TestCheck) Write() {
	c.WriteTested = true
}

var (
	zone         Zone
	fqdnNoDot    = "unit.tests"
	fqdn         = fqdnNoDot + "."
	TypesChecked = map[string]*TestCheck{}
	ipv4         = "192.168.0.1"
	ipv6         = "2601:644:500:e210:62f8:1dff:feb8:947a"
	ipv6local    = "::1"
	randomTxt    = "Bah bah black sheep"
)

func TestMain(m *testing.M) {

	zone = Zone{}
	err := zone.ReadYamlFile("../../testdata/unit.tests.yaml")

	if err != nil {
		fmt.Printf("Error while unmarshalling a zone: %s\n", err.Error())
		os.Exit(1)
	}

	for k, v := range TYPES {
		TypesChecked[k] = &TestCheck{
			Type:        v,
			ReadTested:  false,
			WriteTested: false,
		}
	}

	os.Exit(m.Run())

}

func getType(name string, rtype RType) (record *Record, err error) {

	r, err := zone.FindSubdomain(name)
	if err != nil {
		return nil, fmt.Errorf("zone.FindSubdomain(\"%s\") resulted in error %s", name, err.Error())
	}

	if r.Name != name {
		return nil, fmt.Errorf(`zone.FindSubdomain("a").Name = %q, want %q`, r.Name, name)
	}

	rt, err := r.GetType(rtype.String())
	if err != nil {
		return nil, fmt.Errorf("Error while unmarshalling an a record %s", err.Error())
	}

	if rt.Type != rtype.String() {
		return nil, fmt.Errorf(`zone.GetType("%s").Type = %q, want %q`, rtype.String(), rt.Type, rtype.String())
	}

	return rt, nil

}

func createEmptyType(name string, rtype RType) (record *Record) {

	record = &Record{
		BaseRecord: BaseRecord{
			IsDeleted: false,
			Name:      name,
			Type:      rtype.String(),
			TTL:       300,
			Terraform: Terraform{},
			Octodns:   OctodnsRecordConfig{},
		},
		Values: []RecordValue{},
	}

	return record

}

func checkAmountOfValues(t *testing.T, rt *Record, num int) {

	if len(rt.Values) != num {
		t.Errorf(
			`%s: %s record has %d values,want %d`,
			t.Name(),
			rt.Type,
			len(rt.Values),
			num,
		)
	}
}

func validateStringValues(t *testing.T, rt *Record, wantStrValues []string) {

	strValues := rt.ValuesAsString()
	if strings.Join(strValues, "-") != strings.Join(wantStrValues, "-") {
		t.Errorf(
			`%s ValuesAsString() results in %q, want %q`,
			rt.Type, strValues, wantStrValues,
		)
	}

}

func validateReadSimpleValues(t *testing.T, name string, rType RType, wantValues []string) {
	rt, err := getType(name, rType)
	if err != nil {
		t.Fatal(err.Error())
	}

	check := TypesChecked[rType.String()]
	check.Read()

	checkAmountOfValues(t, rt, len(wantValues))
	validateStringValues(t, rt, wantValues)
}

func validateWriteStringValues(t *testing.T, name string, rType RType, wantValues []string) (rt *Record, err error) {

	rt = createEmptyType(name, rType)

	check := TypesChecked[rType.String()]
	check.Write()

	for _, v := range wantValues {
		err = rt.AddValueFromString(v)
		if err != nil {
			return
		}
	}
	checkAmountOfValues(t, rt, len(wantValues))
	validateStringValues(t, rt, wantValues)
	return
}

func validateWriteWrongStringValues(t *testing.T, name string, rType RType, wrongValues []string) (rt *Record, errValues []string) {
	var err error
	for _, v := range wrongValues {
		rt, err = validateWriteStringValues(t, name, rType, []string{v})
		if err == nil {
			errValues = append(errValues, v)
			t.Errorf("Record %s type accepted wrong value: %q", rt.Type, v)
		}
	}
	return
}

func validateWriteComplexValues(t *testing.T, name string, rType RType, wantsValues []baseRecordValue, wantStrValues []string) (rt *Record, err error) {

	/** TEST **/
	rt, err = validateWriteStringValues(t, name, rType, wantStrValues)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	if !t.Failed() {
		checkAmountOfValues(t, rt, len(wantsValues))

		for i, w := range wantsValues {
			v := rt.Values[i]
			var r DiffReporter
			cmp.Equal(v.baseRecordValue, w, cmp.Reporter(&r))
			if len(r.diffs) > 0 {
				t.Errorf("%s value %d mismatch (-want +got):\n%s", rt.Type, i, r.String())
			}
		}
		validateStringValues(t, rt, wantStrValues)
	}

	return

}

func refFloat64(value float64) *float64 {
	return &value
}
func refInt(value int) *int {
	return &value
}
func refString(value string) *string {
	return &value
}

// DiffReporter is a simple custom reporter that only records differences
// detected during comparison.
type DiffReporter struct {
	path  cmp.Path
	diffs []string
}

func (r *DiffReporter) PushStep(ps cmp.PathStep) {
	r.path = append(r.path, ps)
}

func (r *DiffReporter) Report(rs cmp.Result) {
	if !rs.Equal() {
		vx, vy := r.path.Last().Values()
		r.diffs = append(r.diffs, fmt.Sprintf("%#v:\n\t-: %+v\n\t+: %+v\n", r.path, vx, vy))
	}
}

func (r *DiffReporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

func (r *DiffReporter) String() string {
	return strings.Join(r.diffs, "\n")
}

/****  Simple Record Types: ****/

// TestReadARecord get an A record from unit.tests, checking
// for a valid return value.
func TestReadARecord(t *testing.T) {
	validateReadSimpleValues(t, "www", TYPE_A, []string{"2.2.3.6"})
}

// TestWriteARecord Create an A record, checking
// for a valid return value.
func TestWriteARecord(t *testing.T) {
	var err error
	_, err = validateWriteStringValues(t, "www", TYPE_A, []string{"2.2.3.6", "4.4.4.4"})
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	wrongValues := []string{fqdn, fqdnNoDot, ipv6, ipv6local}
	rt, acceptedValues := validateWriteWrongStringValues(t, "www", TYPE_A, wrongValues)
	_ = rt.Name
	_ = len(acceptedValues)

}

// TestReadAAAARecord get an AAAA record from unit.tests, checking
// for a valid return value.
func TestReadAAAARecord(t *testing.T) {
	validateReadSimpleValues(t, "aaaa", TYPE_AAAA, []string{ipv6})
}

// TestWriteAAAARecord Create an A record, checking
// for a valid return value.
func TestWriteAAAARecord(t *testing.T) {
	//var rt *Record
	var err error
	_, err = validateWriteStringValues(t, TYPE_AAAA.LowerString(), TYPE_AAAA, []string{ipv6})
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	wrongValues := []string{fqdnNoDot, ipv4, randomTxt}
	_, _ = validateWriteWrongStringValues(t, TYPE_AAAA.LowerString(), TYPE_AAAA, wrongValues)

}

// TestReadCNAMERecord get an CNAME record from unit.tests, checking
// for a valid return value.
func TestReadCNAMERecord(t *testing.T) {
	validateReadSimpleValues(t, "cname", TYPE_CNAME, []string{fqdn})
}

// TestWriteCNAMERecord get an CNAME record from unit.tests, checking
// for a valid return value.
func TestWriteCNAMERecord(t *testing.T) {
	var err error
	_, err = validateWriteStringValues(t, TYPE_CNAME.LowerString(), TYPE_CNAME, []string{fqdn})
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	wrongValues := []string{fqdnNoDot, ipv4, ipv6}
	_, _ = validateWriteWrongStringValues(t, TYPE_CNAME.LowerString(), TYPE_CNAME, wrongValues)
}

// TestReadDNAMERecord get an CNAME record from unit.tests, checking
// for a valid return value.
func TestReadDNAMERecord(t *testing.T) {
	validateReadSimpleValues(t, "dname", TYPE_DNAME, []string{fqdn})
}

// TestWriteDNAMERecord get an DNAME record from unit.tests, checking
// for a valid return value.
func TestWriteDNAMERecord(t *testing.T) {
	var err error
	_, err = validateWriteStringValues(t, TYPE_DNAME.LowerString(), TYPE_DNAME, []string{fqdn})
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	wrongValues := []string{fqdnNoDot, ipv4, ipv6}
	_, _ = validateWriteWrongStringValues(t, TYPE_DNAME.LowerString(), TYPE_DNAME, wrongValues)
}

// TestReadPTRRecord get an PTR record from unit.tests, checking
// for a valid return value.
func TestReadPTRRecord(t *testing.T) {
	validateReadSimpleValues(t, "ptr", TYPE_PTR, []string{"foo.bar.com."})
}

// TestWritePTRRecord get an PTR record from unit.tests, checking
// for a valid return value.
func TestWritePTRRecord(t *testing.T) {
	var err error
	_, err = validateWriteStringValues(t, TYPE_PTR.LowerString(), TYPE_PTR, []string{fqdn})
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	wrongValues := []string{fqdnNoDot, ipv4, ipv6, ipv6local}
	_, _ = validateWriteWrongStringValues(t, TYPE_PTR.LowerString(), TYPE_PTR, wrongValues)

}

// TestReadSPFRecord get an SPF record from unit.tests, checking
// for a valid return value.
func TestReadSPFRecord(t *testing.T) {
	validateReadSimpleValues(t, "spf", TYPE_SPF, []string{"v=spf1 ip4:192.168.0.1/16-all"})
}

// TestWriteSPFRecord get an SPF record from unit.tests, checking
// for a valid return value.
func TestWriteSPFRecord(t *testing.T) {
	var err error
	_, err = validateWriteStringValues(t, TYPE_SPF.LowerString(), TYPE_SPF, []string{"v=spf1 ip4:192.168.0.1/16 -all"})
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	wrongValues := []string{fqdnNoDot, ipv4, ipv6, ipv6local, randomTxt}
	_, _ = validateWriteWrongStringValues(t, TYPE_SPF.LowerString(), TYPE_SPF, wrongValues)
}

// TestReadNSRecord get an NS record from unit.tests, checking
// for a valid return value.
func TestReadNSRecord(t *testing.T) {
	validateReadSimpleValues(t, "sub.txt", TYPE_NS, []string{"ns1.test.", "ns2.test."})
}

// TestWriteNSRecord get an NS record from unit.tests, checking
// for a valid return value.
func TestWriteNSRecord(t *testing.T) {
	var err error
	_, err = validateWriteStringValues(t, TYPE_NS.LowerString(), TYPE_NS, []string{"ns1.test.", "ns2.test.", ipv4, ipv6, ipv6local})
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	wrongValues := []string{fqdnNoDot, randomTxt}
	_, _ = validateWriteWrongStringValues(t, TYPE_NS.LowerString(), TYPE_NS, wrongValues)

}

// TestReadTXTRecord get an TXT record from unit.tests, checking
// for a valid return value.
func TestReadTXTRecord(t *testing.T) {
	validateReadSimpleValues(t, "txt", TYPE_TXT, []string{"Bah bah black sheep", "have you any wool.", `v=DKIM1\;k=rsa\;s=email\;h=sha256\;p=A/kinda+of/long/string+with+numb3rs`})

}

// TestWriteTXTRecord get an TXT record from unit.tests, checking
// for a valid return value.
func TestWriteTXTRecord(t *testing.T) {
	var err error
	_, err = validateWriteStringValues(t, TYPE_TXT.LowerString(), TYPE_TXT, []string{fqdnNoDot, ipv4, ipv6, ipv6local, randomTxt})
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	wrongValues := []string{"b;aat"}
	_, _ = validateWriteWrongStringValues(t, TYPE_TXT.LowerString(), TYPE_TXT, wrongValues)

}

/**** Complex Record types ****/

// TestReadMXRecord get an MX record from unit.tests, checking
// for a valid return value.
func TestReadMXRecord(t *testing.T) {

	rt, err := getType("mx", TYPE_MX)
	if err != nil {
		t.Fatal(err.Error())
	}

	check := TypesChecked[TYPE_MX.String()]
	check.Read()

	wants := []baseRecordValue{
		{
			Preference: refInt(40),
			Exchange:   refString("smtp-1." + fqdn),
		},
		{
			Preference: refInt(20),
			Exchange:   refString("smtp-2." + fqdn),
		},
		{
			Preference: refInt(30),
			Exchange:   refString("smtp-3." + fqdn),
		},
		{
			Priority: refInt(10),
			Value:    refString("smtp-4." + fqdn),
		},
	}

	checkAmountOfValues(t, rt, len(wants))

	wantStrValues := []string{}
	for i, w := range wants {

		v := rt.Values[i]

		if w.Preference != nil && w.Exchange != nil {
			wantStrValues = append(
				wantStrValues,
				fmt.Sprintf("%d %s", *w.Preference, *w.Exchange),
			)
		} else {
			wantStrValues = append(
				wantStrValues,
				fmt.Sprintf("%d %s", *w.Priority, *w.Value),
			)
		}
		var r DiffReporter
		cmp.Equal(v.baseRecordValue, w, cmp.Reporter(&r))
		if len(r.diffs) > 0 {
			t.Errorf("%s value %d mismatch (-want +got):\n%s", rt.Type, i, r.String())
		}

	}

	validateStringValues(t, rt, wantStrValues)

}

// TestWriteMXRecord get an MX record from unit.tests, checking
// for a valid return value.
func TestWriteMXRecord(t *testing.T) {

	wants := []baseRecordValue{
		{
			Preference: refInt(40),
			Exchange:   refString("smtp1." + fqdn),
		},
		{
			Preference: refInt(20),
			Exchange:   refString("smtp-2." + fqdn),
		},
		{
			Preference: refInt(30),
			Exchange:   refString("smtp3." + fqdn),
		},
	}
	wantStrValues := []string{}
	for _, w := range wants {
		if w.Preference != nil && w.Exchange != nil {
			wantStrValues = append(
				wantStrValues,
				fmt.Sprintf("%d %s", *w.Preference, *w.Exchange),
			)
		} else {
			wantStrValues = append(
				wantStrValues,
				fmt.Sprintf("%d %s", *w.Priority, *w.Value),
			)
		}
	}

	_, _ = validateWriteComplexValues(t, TYPE_MX.LowerString(), TYPE_MX, wants, wantStrValues)

}

// TestReadCAARecord get an CAA record from unit.tests, checking
// for a valid return value.
func TestReadCAARecord(t *testing.T) {

	rt, err := getType("", TYPE_CAA)
	if err != nil {
		t.Fatal(err.Error())
	}

	check := TypesChecked[TYPE_CAA.String()]
	check.Read()

	wants := []baseRecordValue{
		{
			Flags: refString("0"),
			Tag:   refString("issue"),
			Value: refString("ca." + fqdnNoDot),
		},
	}

	checkAmountOfValues(t, rt, len(wants))

	wantStrValues := []string{}
	for i, w := range wants {

		v := rt.Values[i]

		wantStrValues = append(wantStrValues, fmt.Sprintf("%s %s %s", *w.Flags, *w.Tag, *w.Value))

		var r DiffReporter
		cmp.Equal(v.baseRecordValue, w, cmp.Reporter(&r))
		if len(r.diffs) > 0 {
			t.Errorf("%s value %d mismatch (-want +got):\n%s", rt.Type, i, r.String())
		}

	}

	validateStringValues(t, rt, wantStrValues)

}

// TestWriteCAARecord get an CAA record from unit.tests, checking
// for a valid return value.
func TestWriteCAARecord(t *testing.T) {

	wants := []baseRecordValue{
		{
			Flags: refString("0"),
			Tag:   refString("issue"),
			Value: refString("ca." + fqdnNoDot),
		},
		{
			Flags: refString("0"),
			Tag:   refString("issuewild"),
			Value: refString("ca." + fqdnNoDot + "; policy=ev"),
		},
		{
			Flags: refString("0"),
			Tag:   refString("iodef"),
			Value: refString("mailto:ca@" + fqdnNoDot),
		},
	}
	wantStrValues := []string{}
	for _, w := range wants {
		wantStrValues = append(
			wantStrValues,
			fmt.Sprintf("%s %s %s", *w.Flags, *w.Tag, *w.Value),
		)
	}

	_, _ = validateWriteComplexValues(t, TYPE_CAA.LowerString(), TYPE_CAA, wants, wantStrValues)

}

// TestReadSRVRecord get an SRV record from unit.tests, checking
// for a valid return value.
func TestReadSRVRecord(t *testing.T) {
	rt, err := getType("_srv._tcp", TYPE_SRV)
	if err != nil {
		t.Fatal(err.Error())
	}

	check := TypesChecked[TYPE_SRV.String()]
	check.Read()

	wants := []baseRecordValue{
		{
			Port:     refInt(30),
			Priority: refInt(12),
			Target:   refString("foo-2." + fqdn),
			Weight:   refInt(20),
		},
		{
			Port:     refInt(30),
			Priority: refInt(10),
			Target:   refString("foo-1." + fqdn),
			Weight:   refInt(20),
		},
	}

	checkAmountOfValues(t, rt, len(wants))

	wantStrValues := []string{}
	for i, w := range wants {

		v := rt.Values[i]

		wantStrValues = append(wantStrValues, fmt.Sprintf("%d %d %d %s", *w.Priority, *w.Weight, *w.Port, *w.Target))

		var r DiffReporter
		cmp.Equal(v.baseRecordValue, w, cmp.Reporter(&r))
		if len(r.diffs) > 0 {
			t.Errorf("%s value %d mismatch (-want +got):\n%s", rt.Type, i, r.String())
		}

	}

	validateStringValues(t, rt, wantStrValues)

}

// TestWriteSRVRecord get an SRV record from unit.tests, checking
// for a valid return value.
func TestWriteSRVRecord(t *testing.T) {

	wants := []baseRecordValue{
		{
			Port:     refInt(30),
			Priority: refInt(12),
			Target:   refString("foo-2." + fqdn),
			Weight:   refInt(20),
		},
		{
			Port:     refInt(30),
			Priority: refInt(10),
			Target:   refString("foo-1." + fqdn),
			Weight:   refInt(20),
		},
	}
	wantStrValues := []string{}
	for _, w := range wants {
		wantStrValues = append(
			wantStrValues,
			fmt.Sprintf("%d %d %d %s", *w.Priority, *w.Weight, *w.Port, *w.Target),
		)
	}

	_, _ = validateWriteComplexValues(t, TYPE_SRV.LowerString(), TYPE_SRV, wants, wantStrValues)

}

// TestReadNAPTRRecord get an NAPTR record from unit.tests, checking
// for a valid return value.
func TestReadNAPTRRecord(t *testing.T) {

	rt, err := getType("naptr", TYPE_NAPTR)
	if err != nil {
		t.Fatal(err.Error())
	}

	check := TypesChecked[TYPE_NAPTR.String()]
	check.Read()

	wants := []baseRecordValue{
		{
			Order:       refInt(100),
			Preference:  refInt(100),
			Flags:       refString("U"),
			Service:     refString("SIP+D2U"),
			Regexp:      refString("!^.*$!sip:info@bar.example.com!"),
			Replacement: refString("."),
		},
		{
			Order:       refInt(10),
			Preference:  refInt(100),
			Flags:       refString("S"),
			Service:     refString("SIP+D2U"),
			Regexp:      refString("!^.*$!sip:info@bar.example.com!"),
			Replacement: refString("."),
		},
	}

	checkAmountOfValues(t, rt, len(wants))

	wantStrValues := []string{}
	for i, w := range wants {

		v := rt.Values[i]

		wantStrValues = append(
			wantStrValues,
			fmt.Sprintf("%d %d %q %q %q %s", *w.Order, *w.Preference, *w.Flags, *w.Service, *w.Regexp, *w.Replacement),
		)
		var r DiffReporter
		cmp.Equal(v.baseRecordValue, w, cmp.Reporter(&r))
		if len(r.diffs) > 0 {
			t.Errorf("%s value %d mismatch (-want +got):\n%s", rt.Type, i, r.String())
		}

	}

	validateStringValues(t, rt, wantStrValues)

}

// TestWriteNAPTRRecord get an NAPTR record from unit.tests, checking
// for a valid return value.
func TestWriteNAPTRRecord(t *testing.T) {

	wants := []baseRecordValue{
		{
			Order:       refInt(100),
			Preference:  refInt(200),
			Flags:       refString("U"),
			Service:     refString("SIP+D2U"),
			Regexp:      refString("!^.*$!sip:info@bar.example.com!"),
			Replacement: refString("."),
		},
		{
			Order:       refInt(10),
			Preference:  refInt(100),
			Flags:       refString("S"),
			Service:     refString("SIP+D2U"),
			Regexp:      refString("!^.*$!sip:info@foo.example.com!"),
			Replacement: refString("."),
		},
	}
	wantStrValues := []string{}
	for _, w := range wants {
		wantStrValues = append(
			wantStrValues,
			fmt.Sprintf("%d %d %q %q %q %s", *w.Order, *w.Preference, *w.Flags, *w.Service, *w.Regexp, *w.Replacement),
		)
	}

	_, _ = validateWriteComplexValues(t, TYPE_NAPTR.LowerString(), TYPE_NAPTR, wants, wantStrValues)

}

// TestReadSSHFPRecord get an SSHFP record from unit.tests, checking
// for a valid return value.
func TestReadSSHFPRecord(t *testing.T) {

	rt, err := getType("", TYPE_SSHFP)
	if err != nil {
		t.Fatal(err.Error())
	}

	check := TypesChecked[TYPE_SSHFP.String()]
	check.Read()

	wants := []baseRecordValue{
		{
			Algorithm:       refInt(1), //(0: reserved; 1: RSA; 2: DSA; 3: ECDSA; 4: Ed25519; 6:Ed448)
			Fingerprint:     refString("bf6b6825d2977c511a475bbefb88aad54a92ac73"),
			FingerprintType: refInt(1), //(0: reserved; 1: SHA-1; 2: SHA-256)
		},
		{
			Algorithm:       refInt(1),
			Fingerprint:     refString("7491973e5f8b39d5327cd4e08bc81b05f7710b49"),
			FingerprintType: refInt(1),
		},
	}

	checkAmountOfValues(t, rt, len(wants))

	wantStrValues := []string{}
	for i, w := range wants {
		v := rt.Values[i]

		wantStrValues = append(wantStrValues, fmt.Sprintf("%d %d %s", *w.Algorithm, *w.FingerprintType, *w.Fingerprint))

		var r DiffReporter
		cmp.Equal(v.baseRecordValue, w, cmp.Reporter(&r))
		if len(r.diffs) > 0 {
			t.Errorf("%s value %d mismatch (-want +got):\n%s", rt.Type, i, r.String())
		}
	}

	validateStringValues(t, rt, wantStrValues)

}

// TestWriteSSHFPRecord get an SSHFP record from unit.tests, checking
// for a valid return value.

func TestWriteSSHFPRecord(t *testing.T) {

	wants := []baseRecordValue{
		{
			Algorithm:       refInt(1), //(0: reserved; 1: RSA; 2: DSA; 3: ECDSA; 4: Ed25519; 6:Ed448)
			Fingerprint:     refString("bf6b6825d2977c511a475bbefb88aad54a92ac73"),
			FingerprintType: refInt(1), //(0: reserved; 1: SHA-1; 2: SHA-256)
		},
		{
			Algorithm:       refInt(1),
			Fingerprint:     refString("7491973e5f8b39d5327cd4e08bc81b05f7710b49"),
			FingerprintType: refInt(1),
		},
	}
	wantStrValues := []string{}
	for _, w := range wants {
		wantStrValues = append(
			wantStrValues,
			fmt.Sprintf("%d %d %s", *w.Algorithm, *w.FingerprintType, *w.Fingerprint),
		)
	}

	_, _ = validateWriteComplexValues(t, TYPE_SSHFP.LowerString(), TYPE_SSHFP, wants, wantStrValues)

}

// TestReadURLFWDRecord get an URLFWD record from unit.tests, checking
// for a valid return value.
func TestReadURLFWDRecord(t *testing.T) {

	rt, err := getType("urlfwd", TYPE_URLFWD)
	if err != nil {
		t.Fatal(err.Error())
	}

	check := TypesChecked[TYPE_URLFWD.String()]
	check.Read()
	wants := []baseRecordValue{
		{
			Code:    refInt(302),
			Masking: refInt(2),
			Path:    refString("/"),
			Query:   refInt(0),
			Target:  refString("http://www.unit.tests"),
		},
		{
			Code:    refInt(301),
			Masking: refInt(2),
			Path:    refString("/target"),
			Query:   refInt(0),
			Target:  refString("http://target.unit.tests"),
		},
	}

	checkAmountOfValues(t, rt, len(wants))

	wantStrValues := []string{}
	for i, w := range wants {
		v := rt.Values[i]

		wantStrValues = append(wantStrValues, fmt.Sprintf("%d %d %s %d %s", *w.Code, *w.Masking, *w.Path, *w.Query, *w.Target))

		var r DiffReporter
		cmp.Equal(v.baseRecordValue, w, cmp.Reporter(&r))
		if len(r.diffs) > 0 {
			t.Errorf("%s value %d mismatch (-want +got):\n%s", rt.Type, i, r.String())
		}
	}

	validateStringValues(t, rt, wantStrValues)

}

// TestWriteURLFWDRecord get an URLFWD record from unit.tests, checking
// for a valid return value.
func TestWriteURLFWDRecord(t *testing.T) {

	wants := []baseRecordValue{
		{
			Code:    refInt(302),
			Masking: refInt(2),
			Path:    refString("/"),
			Query:   refInt(0),
			Target:  refString("http://www.unit.tests"),
		},
		{
			Code:    refInt(301),
			Masking: refInt(2),
			Path:    refString("/target"),
			Query:   refInt(0),
			Target:  refString("http://target.unit.tests"),
		},
	}
	wantStrValues := []string{}
	for _, w := range wants {
		wantStrValues = append(
			wantStrValues,
			fmt.Sprintf("%d %d %s %d %s", *w.Code, *w.Masking, *w.Path, *w.Query, *w.Target),
		)
	}

	_, _ = validateWriteComplexValues(t, TYPE_URLFWD.LowerString(), TYPE_URLFWD, wants, wantStrValues)

}

// TestReadLOCRecord get an LOC record from unit.tests, checking
// for a valid return value.
func TestReadLOCRecord(t *testing.T) {

	rt, err := getType("loc", TYPE_LOC)
	if err != nil {
		t.Fatal(err.Error())
	}

	check := TypesChecked[TYPE_LOC.String()]
	check.Read()

	wants := []baseRecordValue{
		{
			Altitude:      refFloat64(20),
			LatDegrees:    refInt(31),
			LatDirection:  refString("S"),
			LatMinutes:    refInt(58),
			LatSeconds:    refFloat64(52.1),
			LongDegrees:   refInt(115),
			LongDirection: refString("E"),
			LongMinutes:   refInt(49),
			LongSeconds:   refFloat64(11.7),
			PrecisionHorz: refInt(10),
			PrecisionVert: refInt(2),
			Size:          refInt(10),
		},
		{
			Altitude:      refFloat64(20),
			LatDegrees:    refInt(53),
			LatDirection:  refString("N"),
			LatMinutes:    refInt(13),
			LatSeconds:    refFloat64(10),
			LongDegrees:   refInt(2),
			LongDirection: refString("W"),
			LongMinutes:   refInt(18),
			LongSeconds:   refFloat64(26),
			PrecisionHorz: refInt(1000),
			PrecisionVert: refInt(2),
			Size:          refInt(10),
		},
	}

	checkAmountOfValues(t, rt, len(wants))

	wantStrValues := []string{}
	for i, w := range wants {
		v := rt.Values[i]

		wantStrValues = append(wantStrValues,
			fmt.Sprintf(
				"%d %d %0.2f %s %d %d %0.2f %s %0.2f %d %d %d",
				*w.LatDegrees,
				*w.LatMinutes,
				*w.LatSeconds,
				*w.LatDirection,
				*w.LongDegrees,
				*w.LongMinutes,
				*w.LongSeconds,
				*w.LongDirection,
				*w.Altitude,
				*w.Size,
				*w.PrecisionHorz,
				*w.PrecisionVert,
			),
		)

		var r DiffReporter
		cmp.Equal(v.baseRecordValue, w, cmp.Reporter(&r))
		if len(r.diffs) > 0 {
			t.Errorf("%s value %d mismatch (-want +got):\n%s", rt.Type, i, r.String())
		}
	}

	validateStringValues(t, rt, wantStrValues)

}

// TestWriteLOCRecord get an LOC record from unit.tests, checking
// for a valid return value.
func TestWriteLOCRecord(t *testing.T) {
	wants := []baseRecordValue{
		{
			LatDegrees:   refInt(31),
			LatDirection: refString("S"),
			LatMinutes:   refInt(58),
			LatSeconds:   refFloat64(52.1),

			LongDegrees:   refInt(115),
			LongDirection: refString("E"),
			LongMinutes:   refInt(49),
			LongSeconds:   refFloat64(11.7),

			Altitude: refFloat64(20),

			Size:          refInt(10),
			PrecisionHorz: refInt(10),
			PrecisionVert: refInt(2),
		},
		{
			LatDegrees:   refInt(53),
			LatDirection: refString("N"),
			LatMinutes:   refInt(13),
			LatSeconds:   refFloat64(10),

			LongDegrees:   refInt(2),
			LongDirection: refString("W"),
			LongMinutes:   refInt(18),
			LongSeconds:   refFloat64(26),

			Altitude: refFloat64(20),

			Size:          refInt(10),
			PrecisionHorz: refInt(1000),
			PrecisionVert: refInt(2),
		},
	}
	wantStrValues := []string{}
	for _, w := range wants {
		wantStrValues = append(
			wantStrValues,
			fmt.Sprintf(
				"%d %d %0.2f %s %d %d %0.2f %s %0.2f %d %d %d",
				*w.LatDegrees,
				*w.LatMinutes,
				*w.LatSeconds,
				*w.LatDirection,
				*w.LongDegrees,
				*w.LongMinutes,
				*w.LongSeconds,
				*w.LongDirection,
				*w.Altitude,
				*w.Size,
				*w.PrecisionHorz,
				*w.PrecisionVert,
			),
		)
	}

	_, _ = validateWriteComplexValues(t, TYPE_LOC.LowerString(), TYPE_LOC, wants, wantStrValues)

}

/**** Other Tests ****/
// TestReadAllChecks check result of all tests.
func TestReadAllChecks(t *testing.T) {

	for _, c := range TypesChecked {

		if c.ReadTested == false {
			t.Errorf("No Read test logged for Type %q", c.Type.String())
		}
	}

	for _, c := range TypesChecked {

		if c.WriteTested == false {
			t.Errorf("No Write test logged for Type %q", c.Type.String())
		}

	}

}
