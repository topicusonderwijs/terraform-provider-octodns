// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

/*
func TestAccRecordDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccRecordDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.octodns_record.test", "id", "external-topicus.education-"),
				),
			},
		},
	})
}

const testAccRecordDataSourceConfig = `
data "octodns_a_record" "test" {
  id="external-topicus.education-"
  name="blaat"
  scope="external"
  zone="example.com"
}
`
*/
