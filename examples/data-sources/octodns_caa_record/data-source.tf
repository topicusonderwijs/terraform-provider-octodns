data "octodns_caa_record" "root" {
  zone = "unit.tests"
  name = "@"
}
output "caa_record" {
  value = data.octodns_caa_record.root
}


