locals {

  caa_raw = [
    { flags = 30, tag = "issue", value = "ca.unit.tests" },
  ]

  caa_values = [for v in local.caa_raw : "${v.flags} ${v.tag} ${v.value}"]

}

resource "octodns_caa_record" "root" {
  zone   = "example.com"
  name   = "@"
  ttl    = 300
  values = local.caa_values
}


