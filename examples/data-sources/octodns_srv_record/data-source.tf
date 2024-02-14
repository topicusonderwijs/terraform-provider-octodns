data "octodns_srv_record" "root" {
  zone = "unit.tests"
  name = "_imap._tcp"
}
output "srv_record" {
  value = data.octodns_srv_record.root
}


