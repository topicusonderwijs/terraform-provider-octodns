// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/topicusonderwijs/terraform-provider-octodns/internal/models"
)

func TestCoexistenceWarning(t *testing.T) {

	cases := []struct {
		name      string
		rtype     string
		conflicts []string
		contains  []string
	}{
		{
			"ALIAS next to A and AAAA",
			"ALIAS", []string{"A", "AAAA"},
			[]string{"existing A and AAAA record(s)", "ALIAS records cannot coexist with A or AAAA records"},
		},
		{
			"CNAME next to three types",
			"CNAME", []string{"A", "MX", "TXT"},
			[]string{"existing A, MX and TXT record(s)", "CNAME records cannot coexist with other records"},
		},
		{
			"A next to ALIAS and CNAME",
			"A", []string{"ALIAS", "CNAME"},
			[]string{"existing ALIAS and CNAME record(s)", "CNAME records cannot coexist with other records; ALIAS records cannot coexist with A or AAAA records"},
		},
		{
			"TXT next to CNAME",
			"TXT", []string{"CNAME"},
			[]string{"existing CNAME record(s)", "CNAME records cannot coexist with other records"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := coexistenceWarning(c.rtype, "@", "example.com", c.conflicts)
			for _, want := range c.contains {
				if !strings.Contains(got, want) {
					t.Errorf("warning %q does not contain %q", got, want)
				}
			}
		})
	}
}

func TestIsMissingRecordError(t *testing.T) {

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"subdomain not found", models.ErrSubdomainNotFound, true},
		{"type not found", models.ErrTypeNotFound, true},
		{"wrapped type not found", fmt.Errorf("type 'A' not found: %w", models.ErrTypeNotFound), true},
		{"other error", errors.New("zone.doc is not a document node"), false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isMissingRecordError(c.err); got != c.want {
				t.Errorf("isMissingRecordError(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}
}

func TestIgnoreMissingRecord(t *testing.T) {

	for _, errorOnMissing := range []bool{false, true} {
		r := &RecordResource{client: &models.GitHubClient{ErrorOnMissingRecords: errorOnMissing}}

		if got := r.ignoreMissingRecord(models.ErrSubdomainNotFound); got == errorOnMissing {
			t.Errorf("error_on_missing_records=%v: ignoreMissingRecord(ErrSubdomainNotFound) = %v", errorOnMissing, got)
		}
		if r.ignoreMissingRecord(errors.New("other error")) {
			t.Errorf("error_on_missing_records=%v: other errors must never be ignored", errorOnMissing)
		}
	}
}
