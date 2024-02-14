resource "octodns_dname_record" "dname" {
  zone   = "example.com"
  name   = "dname"
  values = ["unit.tests."]
}


