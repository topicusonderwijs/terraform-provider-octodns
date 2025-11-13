package models

import (
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
