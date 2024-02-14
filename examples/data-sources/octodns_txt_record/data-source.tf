data "octodns_txt_record" "txt" {
  zone = "unit.tests"
  name = "txt"
}
output "txt_record" {
  value = data.octodns_txt_record.txt
}


