data "octodns_dname_record" "dname" {
  zone = "unit.tests"
  name = "dname"
}
output "dname_record" {
  value = data.octodns_dname_record.dname
}


