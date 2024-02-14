data "octodns_naptr_record" "naptr" {
  zone = "unit.tests"
  name = "naptr"
}
output "naptr_record" {
  value = data.octodns_naptr_record.naptr
}


