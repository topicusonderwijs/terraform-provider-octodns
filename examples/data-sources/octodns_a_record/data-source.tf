data "octodns_a_record" "root" {
  zone = "unit.tests"
  name = "@"
}
output "a_record" {
  value = data.octodns_a_record.root
}


