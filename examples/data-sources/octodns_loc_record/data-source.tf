data "octodns_loc_record" "loc" {
  zone = "unit.tests"
  name = "loc"
}
output "loc_record" {
  value = data.octodns_loc_record.loc
}


