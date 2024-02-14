data "octodns_urlfwd_record" "urlfwd" {
  zone = "unit.tests"
  name = "urlfwd"
}
output "urlfwd_record" {
  value = data.octodns_urlfwd_record.urlfwd
}












