data "octodns_cname_record" "cname" {
  zone = "unit.tests"
  name = "cname"
}
output "cname_record" {
  value = data.octodns_cname_record.cname
}


