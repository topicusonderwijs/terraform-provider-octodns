data "octodns_aaaa_record" "aaaa" {
  zone = "unit.tests"
  name = "aaaa"
}
output "aaaa_record" {
  value = data.octodns_aaaa_record.aaaa
}


