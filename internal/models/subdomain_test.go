package models

import (
	"bytes"
	"errors"
	"testing"
)

func TestSubdomain_UpdateYaml(t *testing.T) {

	xZone, err := zoneFromYaml(UNIT_FILE_DEFAULT)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	sub, err := xZone.FindSubdomain("cname")
	if err != nil {
		t.Errorf("FindSubdomain throws an error: %s", err)
	}

	err = sub.FindAllType()
	if err != nil {
		t.Errorf("FindAllTypes throws an error: %s", err)
	}

	err = sub.UpdateYaml()
	if err != nil {
		t.Errorf("UpdateYaml throws an error: %s", err)
	}

	compareZoneOutputWithFile(t, &xZone, UNIT_FILE_DEFAULT)

	//	fmt.Printf("%s", out)

}

func TestSubdomain_DeleteType(t *testing.T) {

	xZone, err := zoneFromYaml(UNIT_FILE_LOC_DELETED)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	sub, err := xZone.FindSubdomain("")
	if err != nil {
		t.Errorf("FindSubdomain throws an error: %s", err)
	}

	err = sub.FindAllType()
	if err != nil {
		t.Errorf("FindAllType throws an error: %s", err)
	}

	_, err = sub.GetType(TYPE_NS.String())
	if err != nil {
		t.Errorf("GetType throws an error: %s", err)
	}

	err = sub.DeleteType(TYPE_NS.String())
	if err != nil {
		t.Errorf("DeleteType throws an error: %s", err)
	}

	_, err = sub.GetType(TYPE_NS.String())
	if err != nil && !errors.Is(err, ErrTypeNotFound) {
		t.Errorf("DeleteType->GetType throws an error: %s", err)
	} else if err == nil {
		t.Errorf("DeleteType Type still present after deleting type")
	}

	// Delete mapping node

	sub, err = xZone.FindSubdomain("aaaa")
	if err != nil {
		t.Errorf("FindSubdomain1 throws an error: %s", err)
	}

	// Delete non-existing node

	err = sub.DeleteType(TYPE_A.String())
	if err == nil {
		t.Errorf("DeleteType should throw an error for non-existing type")
	}

	err = sub.DeleteType(TYPE_AAAA.String())
	if err != nil {
		t.Errorf("DeleteType throws an error: %s", err)
	}

	sub, err = xZone.FindSubdomain("caa")
	if err != nil && !errors.Is(err, ErrSubdomainNotFound) {
		t.Errorf("FindSubdomain2 throws an error: %s", err)
	} else if err == nil {
		t.Errorf("DeleteType Type still present after deleting type")
	}

}

func TestSubdomain_GetType(t *testing.T) {

	xZone, err := zoneFromYaml(UNIT_FILE_LOC_DELETED)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	sub, err := xZone.FindSubdomain("")
	if err != nil {
		t.Errorf("FindSubdomain throws an error: %s", err)
	}

	_, err = sub.GetType("Invalid")
	if err == nil {
		t.Errorf("GetType should throw an error for invalid type name")
	}

	_, err = sub.GetType(TYPE_NS.String())
	if err != nil {
		t.Errorf("GetType throws an error: %s", err)
	}

}

func TestSubdomain_RestoreValuesRollback(t *testing.T) {

	xZone, err := zoneFromYaml(UNIT_FILE_DEFAULT)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	sub, err := xZone.FindSubdomain("txt")
	if err != nil {
		t.Errorf("FindSubdomain throws an error: %s", err)
	}

	err = sub.FindAllType()
	if err != nil {
		t.Errorf("FindAllType throws an error: %s", err)
	}

	record, err := sub.GetType(TYPE_TXT.String())
	if err != nil {
		t.Errorf("GetType throws an error: %s", err)
	}

	// Snapshot original state (same pattern as Update rollback)
	oldValues := make([]RecordValue, len(record.Values))
	copy(oldValues, record.Values)
	oldTTL := record.TTL
	originalValueCount := len(record.Values)

	// Capture baseline YAML after first UpdateYaml (before any modification)
	if err = sub.UpdateYaml(); err != nil {
		t.Errorf("UpdateYaml throws an error: %s", err)
	}
	baseline, err := xZone.WriteYaml()
	if err != nil {
		t.Errorf("WriteYaml throws an error: %s", err)
	}

	// Simulate a failed RecordFromDataModel: clear and replace values, change TTL
	record.ClearValues()
	if err = record.AddValueFromString("simulated new value"); err != nil {
		t.Errorf("AddValueFromString throws an error: %s", err)
	}
	record.TTL = 1234

	if err = sub.UpdateYaml(); err != nil {
		t.Errorf("UpdateYaml throws an error: %s", err)
	}

	modified, err := xZone.WriteYaml()
	if err != nil {
		t.Errorf("WriteYaml throws an error: %s", err)
	}
	if bytes.Equal(baseline, modified) {
		t.Errorf("YAML did not change after modification — test setup failed")
	}

	// Rollback
	record.Values = oldValues
	record.TTL = oldTTL
	if err = sub.UpdateYaml(); err != nil {
		t.Errorf("UpdateYaml after rollback throws an error: %s", err)
	}

	// Record state must match the snapshot
	if record.TTL != oldTTL {
		t.Errorf("TTL not restored: got %d, want %d", record.TTL, oldTTL)
	}
	if len(record.Values) != originalValueCount {
		t.Errorf("Value count not restored: got %d, want %d", len(record.Values), originalValueCount)
	}

	// YAML must match the baseline (same as before modification)
	restored, err := xZone.WriteYaml()
	if err != nil {
		t.Errorf("WriteYaml throws an error: %s", err)
	}
	if !bytes.Equal(baseline, restored) {
		t.Errorf("YAML not restored to baseline after rollback")
	}
}

func TestSubdomain_CreateRollbackNewSubdomain(t *testing.T) {

	xZone, err := zoneFromYaml(UNIT_FILE_DEFAULT)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	baseline, err := xZone.WriteYaml()
	if err != nil {
		t.Errorf("WriteYaml throws an error: %s", err)
	}

	// Simulate Create: new subdomain + new type + value
	sub, err := xZone.CreateSubdomain("rollback-test")
	if err != nil {
		t.Errorf("CreateSubdomain throws an error: %s", err)
	}

	record, err := sub.CreateType(TYPE_A.String())
	if err != nil {
		t.Errorf("CreateType throws an error: %s", err)
	}

	if err = record.AddValueFromString("192.168.0.1"); err != nil {
		t.Errorf("AddValueFromString throws an error: %s", err)
	}

	if err = sub.UpdateYaml(); err != nil {
		t.Errorf("UpdateYaml throws an error: %s", err)
	}

	modified, err := xZone.WriteYaml()
	if err != nil {
		t.Errorf("WriteYaml throws an error: %s", err)
	}
	if bytes.Equal(baseline, modified) {
		t.Errorf("YAML did not change after modification — test setup failed")
	}

	// Rollback: subdomain was newly created, delete the whole subdomain
	if err = xZone.DeleteSubdomain(sub.Name); err != nil {
		t.Errorf("DeleteSubdomain throws an error: %s", err)
	}

	restored, err := xZone.WriteYaml()
	if err != nil {
		t.Errorf("WriteYaml throws an error: %s", err)
	}
	if !bytes.Equal(baseline, restored) {
		t.Errorf("YAML not restored to baseline after rollback of new subdomain")
	}
}

func TestSubdomain_CreateRollbackExistingSubdomain(t *testing.T) {

	xZone, err := zoneFromYaml(UNIT_FILE_DEFAULT)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// Apex ('') already has multiple types (A, SSHFP, NS, CAA); add MX and roll back
	sub, err := xZone.FindSubdomain("")
	if err != nil {
		t.Errorf("FindSubdomain throws an error: %s", err)
	}
	if err = sub.FindAllType(); err != nil {
		t.Errorf("FindAllType throws an error: %s", err)
	}

	// Baseline YAML must be taken AFTER FindAllType so node state matches post-rollback
	if err = sub.UpdateYaml(); err != nil {
		t.Errorf("UpdateYaml throws an error: %s", err)
	}
	baseline, err := xZone.WriteYaml()
	if err != nil {
		t.Errorf("WriteYaml throws an error: %s", err)
	}

	record, err := sub.CreateType(TYPE_MX.String())
	if err != nil {
		t.Errorf("CreateType throws an error: %s", err)
	}
	if err = record.AddValueFromString("10 mail.unit.tests."); err != nil {
		t.Errorf("AddValueFromString throws an error: %s", err)
	}
	if err = sub.UpdateYaml(); err != nil {
		t.Errorf("UpdateYaml throws an error: %s", err)
	}

	modified, err := xZone.WriteYaml()
	if err != nil {
		t.Errorf("WriteYaml throws an error: %s", err)
	}
	if bytes.Equal(baseline, modified) {
		t.Errorf("YAML did not change after modification — test setup failed")
	}

	// Rollback: subdomain existed, delete only the added type
	if err = sub.DeleteType(TYPE_MX.String()); err != nil {
		t.Errorf("DeleteType throws an error: %s", err)
	}
	if err = sub.UpdateYaml(); err != nil {
		t.Errorf("UpdateYaml after rollback throws an error: %s", err)
	}

	restored, err := xZone.WriteYaml()
	if err != nil {
		t.Errorf("WriteYaml throws an error: %s", err)
	}
	if !bytes.Equal(baseline, restored) {
		t.Errorf("YAML not restored to baseline after rollback of added type")
	}
}

func TestSubdomain_CreateType(t *testing.T) {

	xZone, err := zoneFromYaml(UNIT_FILE_LOC_DELETED)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	sub, err := xZone.FindSubdomain("aaaa")
	if err != nil {
		t.Errorf("FindSubdomain throws an error: %s", err)
	}

	_, err = sub.CreateType(TYPE_TXT.String())
	if err != nil {
		t.Errorf("CreateType throws an error: %s", err)
	}

	_, err = sub.CreateType(TYPE_AAAA.String())
	if err == nil {
		t.Errorf("CreateType should throw an error for existing type")
	}

}
