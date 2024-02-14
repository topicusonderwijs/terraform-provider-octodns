data "octodns_sshfp_record" "root" {
  zone = "unit.tests"
  name = "@"
}
output "sshfp_record" {
  value = data.octodns_sshfp_record.root
}


