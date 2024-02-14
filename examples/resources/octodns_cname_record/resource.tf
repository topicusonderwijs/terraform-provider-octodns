resource "octodns_cname_record" "cname" {
  zone   = "example.com"
  name   = "cname"
  values = ["unit.tests."]
}


