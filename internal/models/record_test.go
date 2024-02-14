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

func checkAmountOfValues(t *testing.T, rt *Record, num int) {

	if len(rt.Values) != num {
		t.Fatalf(
			`%s record has %d values,want %d`,
			rt.Type,
			len(rt.Values),
			num,
		)
	}
}

func validateStringValues(t *testing.T, rt *Record, wantStrValues []string) {

	strValues := rt.ValuesAsString()
	if strings.Join(strValues, "-") != strings.Join(wantStrValues, "-") {
		t.Fatalf(
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

/****  Simple Record Types: ****/

// TestARecord get an A record from unit.tests, checking
// for a valid return value.
func TestARecord(t *testing.T) {
	validateReadSimpleValues(t, "www", TYPE_A, []string{"2.2.3.6"})
}

// TestAAAARecord get an AAAA record from unit.tests, checking
// for a valid return value.
func TestAAAARecord(t *testing.T) {
	validateReadSimpleValues(t, "aaaa", TYPE_AAAA, []string{"2601:644:500:e210:62f8:1dff:feb8:947a"})
}

// TestCNAMERecord get an CNAME record from unit.tests, checking
// for a valid return value.
func TestCNAMERecord(t *testing.T) {
	validateReadSimpleValues(t, "cname", TYPE_CNAME, []string{fqdn})
}

// TestDNAMERecord get an CNAME record from unit.tests, checking
// for a valid return value.
func TestDNAMERecord(t *testing.T) {
	validateReadSimpleValues(t, "dname", TYPE_DNAME, []string{fqdn})
}

// TestPTRRecord get an PTR record from unit.tests, checking
// for a valid return value.
func TestPTRRecord(t *testing.T) {
	validateReadSimpleValues(t, "ptr", TYPE_PTR, []string{"foo.bar.com."})
}

// TestSPFRecord get an SPF record from unit.tests, checking
// for a valid return value.
func TestSPFRecord(t *testing.T) {
	validateReadSimpleValues(t, "spf", TYPE_SPF, []string{"v=spf1 ip4:192.168.0.1/16-all"})
}

// TestNSRecord get an NS record from unit.tests, checking
// for a valid return value.
func TestNSRecord(t *testing.T) {
	validateReadSimpleValues(t, "sub.txt", TYPE_NS, []string{"ns1.test.", "ns2.test."})
}

// TestTXTRecord get an TXT record from unit.tests, checking
// for a valid return value.
func TestTXTRecord(t *testing.T) {
	validateReadSimpleValues(t, "txt", TYPE_TXT, []string{"Bah bah black sheep", "have you any wool.", `v=DKIM1\;k=rsa\;s=email\;h=sha256\;p=A/kinda+of/long/string+with+numb3rs`})

}

/**** Complex Record types ****/

// TestMXRecord get an MX record from unit.tests, checking
// for a valid return value.
func TestMXRecord(t *testing.T) {

	rt, err := getType("mx", TYPE_MX)
	if err != nil {
		t.Fatal(err.Error())
	}

	check := TypesChecked[TYPE_MX.String()]
	check.Read()

	wants := []struct {
		E string
		P int
		L bool
	}{
		{E: "smtp-1." + fqdn, P: 40, L: false},
		{E: "smtp-2." + fqdn, P: 20, L: false},
		{E: "smtp-3." + fqdn, P: 30, L: false},
		{E: "smtp-4." + fqdn, P: 10, L: true},
	}

	checkAmountOfValues(t, rt, len(wants))

	wantStrValues := []string{}
	for i, w := range wants {

		wantStrValues = append(wantStrValues, fmt.Sprintf("%d %s", w.P, w.E))

		v := rt.Values[i]
		if w.L {

			if *v.Value != w.E || *v.Priority != w.P {
				t.Fatalf(
					`%s value %d has Value = %q, Priority = %q, want %q / %q`,
					rt.Type, i,
					*v.Value, *v.Priority,
					w.E, w.P,
				)
			}
		} else {
			if *v.Exchange != w.E || *v.Preference != w.P {
				t.Fatalf(
					`%s value %d Exchange = %q, Preference = %q, want %q / %q`,
					rt.Type, i,
					*v.Exchange, *v.Preference,
					w.E, w.P,
				)
			}
		}
	}

	validateStringValues(t, rt, wantStrValues)

}

// TestCAARecord get an CAA record from unit.tests, checking
// for a valid return value.
func TestCAARecord(t *testing.T) {

	rt, err := getType("", TYPE_CAA)
	if err != nil {
		t.Fatal(err.Error())
	}

	check := TypesChecked[TYPE_CAA.String()]
	check.Read()

	wants := []struct {
		F string
		T string
		V string
	}{
		{F: "0", T: "issue", V: "ca." + fqdnNoDot},
	}

	checkAmountOfValues(t, rt, len(wants))

	wantStrValues := []string{}
	for i, w := range wants {

		v := rt.Values[i]

		wantStrValues = append(wantStrValues, fmt.Sprintf("%s %s %s", w.F, w.T, w.V))

		if *v.Flags != w.F || *v.Tag != w.T || *v.Value != w.V {
			t.Fatalf(
				`%s value %d has Flags = %q, Tag = %q, Value = %q, want %q / %q / %q`,
				rt.Type, i,
				*v.Flags, *v.Tag, *v.Value,
				w.F, w.T, w.V,
			)
		}

	}

	validateStringValues(t, rt, wantStrValues)

}

// TestSRVRecord get an SRV record from unit.tests, checking
// for a valid return value.
func TestSRVRecord(t *testing.T) {
	rt, err := getType("_imap._tcp", TYPE_SRV)
	if err != nil {
		t.Fatal(err.Error())
	}

	check := TypesChecked[TYPE_SRV.String()]
	check.Read()

	wants := []struct {
		Po int
		Pr int
		T  string
		W  int
	}{
		{Po: 0, Pr: 0, T: ".", W: 0},
	}

	checkAmountOfValues(t, rt, len(wants))

	wantStrValues := []string{}
	for i, w := range wants {

		v := rt.Values[i]

		wantStrValues = append(wantStrValues, fmt.Sprintf("%d %d %d %s", w.Po, w.Pr, w.W, w.T))

		if *v.Port != w.Po || *v.Priority != w.Pr || *v.Target != w.T || *v.Weight != w.W {
			t.Fatalf(
				`%s value %d has Port = %q, Priority = %q, Target = %q, Weight = %q, want %q / %q / %q / %q`,
				rt.Type, i,
				*v.Port, *v.Priority, *v.Target, *v.Weight,
				w.Po, w.Pr, w.T, w.W,
			)
		}

	}

	validateStringValues(t, rt, wantStrValues)

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

// TestNAPTRRecord get an NAPTR record from unit.tests, checking
// for a valid return value.
func TestNAPTRRecord(t *testing.T) {

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

// TestSSHFPRecord get an SSHFP record from unit.tests, checking
// for a valid return value.
func TestSSHFPRecord(t *testing.T) {

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

// TestURLFWDRecord get an URLFWD record from unit.tests, checking
// for a valid return value.
func TestURLFWDRecord(t *testing.T) {

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

// TestLOCRecord get an LOC record from unit.tests, checking
// for a valid return value.
func TestLOCRecord(t *testing.T) {

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

/**** Other Tests ****/
// TestAllChecks check result of all tests
func TestAllChecks(t *testing.T) {

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
