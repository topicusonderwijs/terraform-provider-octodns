data "octodns_mx_record" "mx" {
  zone = "unit.tests"
  name = "mx"
}
output "mx_record" {
  value = data.octodns_mx_record.mx
}


