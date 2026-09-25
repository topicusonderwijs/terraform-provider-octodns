// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"strings"
	"testing"
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
