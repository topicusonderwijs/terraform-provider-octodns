package models

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"os"
	"testing"
)

const (
	UNIT_FILE_PATH        = "../../testdata/"
	UNIT_FILE_DEFAULT     = "unit.tests.yaml"
	UNIT_FILE_LOC_DELETED = "unit.tests.loc_deleted.yaml"
	UNIT_FILE_UNIT_ADDED  = "unit.tests.unit_added.yaml"
)

func zoneFromYaml(filename string) (zone Zone, err error) {
	zone = Zone{}
	err = zone.ReadYamlFile(UNIT_FILE_PATH + filename)
	if err != nil {
		err = fmt.Errorf("Error while unmarshalling a zone: %s\n", err.Error())
		return
	}

	return
}

func readYamlUnitFile(filename string) (fileContent []byte, err error) {

	fileContent, err = os.ReadFile(UNIT_FILE_PATH + filename)
	if err != nil {
		err = fmt.Errorf("could not read yaml file for verification: %s", err.Error())
		return
	}

	if bytes.Contains(fileContent, []byte("---\n? ''\n:")) {
		// Unit test has odd `? ''` start syntax, replacing it with the output format of yamlv3
		fileContent = bytes.Replace(fileContent, []byte("---\n? ''\n:"), []byte("'':\n "), 1)
	}

	return
}

func compareZoneOutputWithFile(t *testing.T, zone *Zone, filename string) {

	out, err := zone.WriteYaml()
	if err != nil {
		t.Errorf("could not write zone to yaml: %s", err.Error())
		return
	}

	fileContent, err := readYamlUnitFile(filename)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	diff := cmp.Diff(fileContent, out)

	if diff != "" {
		t.Errorf("Output is not equal:\n%s", diff)
	}

}

func TestZone_ReadYamlFile(t *testing.T) {

	tmpZone := Zone{}
	err := tmpZone.ReadYamlFile(UNIT_FILE_PATH + UNIT_FILE_DEFAULT)
	if err != nil {
		t.Errorf("Error while loading existing yaml: %s", err.Error())
	}

	err = tmpZone.ReadYamlFile(UNIT_FILE_PATH + "rem_" + UNIT_FILE_DEFAULT)
	if err == nil {
		t.Errorf("ReadYamlFile did not throw an error for a non existing file")
	}

}

func TestZone_ReadYaml(t *testing.T) {

	tmpZone := Zone{}
	err := tmpZone.ReadYaml([]byte("---\n? ''\n  :\n  - ttl: 600\n    type: NS\n    values:\n      - 172.0.0.1"))
	if err != nil {
		t.Errorf("Error while loading test yaml: %s", err.Error())
	}

	err = tmpZone.ReadYaml([]byte("---\nblaat: true"))
	if err != nil {
		t.Errorf("Error while loading valid but bogus yaml: %s", err.Error())
	}
	_, err = tmpZone.WriteYaml()
	if err != nil {
		t.Errorf("Error while loading valid but bogus yaml: %s", err.Error())
	}

}

func TestZone_WriteYaml(t *testing.T) {

	compareZoneOutputWithFile(t, &zone, UNIT_FILE_DEFAULT)

}

func TestZone_FindRecordByType(t *testing.T) {

	recordChild, recordNode, recordParent, err := zone.FindRecordByType("www", TYPE_A.String())
	if err != nil {
		t.Errorf("FindRecordByType error: %s", err.Error())
	}

	if recordChild == nil {
		t.Errorf("FindRecordByType::recordChild returned as nil")
	}
	if recordNode == nil {
		t.Errorf("FindRecordByType::recordChild returned as nil")
	}
	if recordParent == nil {
		t.Errorf("FindRecordByType::recordChild returned as nil")
	}

	recordChild, recordNode, recordParent, err = zone.FindRecordByType("", TYPE_NS.String())
	if err != nil {
		t.Errorf("FindRecordByType error: %s", err.Error())
	}

	if recordChild == nil {
		t.Errorf("FindRecordByType::recordChild returned as nil")
	}
	if recordNode == nil {
		t.Errorf("FindRecordByType::recordChild returned as nil")
	}
	if recordParent == nil {
		t.Errorf("FindRecordByType::recordChild returned as nil")
	}

}

func TestZone_GetRecord(t *testing.T) {

	wantName := "www"
	wantType := TYPE_A.String()

	rec, err := zone.GetRecord(wantName, wantType)
	if err != nil {
		t.Errorf("GetRecord error: %s", err.Error())
	}

	if rec.Name != wantName {
		t.Errorf("recordname mismats, got %q, want %q", rec.Name, wantName)
	}

	if rec.Type != wantType {
		t.Errorf("recordname mismats, got %q, want %q", rec.Name, wantName)
	}

}

func TestZone_DeleteSubdomain(t *testing.T) {

	err := zone.DeleteSubdomain("loc")
	if err != nil {
		t.Errorf("DeleteSubdomain error: %s", err.Error())
	}

	compareZoneOutputWithFile(t, &zone, UNIT_FILE_LOC_DELETED)

}

func TestZone_CreateSubdomain(t *testing.T) {

	testZone, err := zoneFromYaml(UNIT_FILE_DEFAULT)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	s, err := testZone.CreateSubdomain("unit")
	if err != nil {
		t.Errorf("CreateSubdomain error: %s", err.Error())
	}

	rt, err := s.CreateType(TYPE_A.String())
	if err != nil {
		t.Errorf("CreateSubdomain::CreateType error: %s", err.Error())
	}

	rt.TTL = 300

	err = rt.AddValueFromString("127.0.0.1")
	if err != nil {
		t.Errorf("CreateSubdomain::CreateType::AddValueFromString error: %s", err.Error())
	}
	err = rt.AddValueFromString("127.0.0.2")
	if err != nil {
		t.Errorf("CreateSubdomain::CreateType::AddValueFromString error: %s", err.Error())
	}

	_ = rt.UpdateYaml()

	compareZoneOutputWithFile(t, &testZone, UNIT_FILE_UNIT_ADDED)

	s, err = testZone.CreateSubdomain("unit")
	if err == nil {
		t.Errorf("CreateSubdomain did not throw error while trying to create an existing subdomain")
	} else if !errors.Is(err, ErrSubdomainAlreadyExists) {
		t.Errorf("CreateSubdomain unexpected error while trying to create an existing subdomain: %s", err)
	}

}
