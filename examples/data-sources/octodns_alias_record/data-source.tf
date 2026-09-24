data "octodns_alias_record" "root" {
  zone = "unit.tests"
  name = "@"
}
output "alias_record" {
  value = data.octodns_alias_record.root
}
