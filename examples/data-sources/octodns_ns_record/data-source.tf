data "octodns_ns_record" "root" {
  zone = "unit.tests"
  name = "@"
}
output "ns_record" {
  value = data.octodns_ns_record.root
}


