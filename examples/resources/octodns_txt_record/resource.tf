resource "octodns_txt_record" "minimal" {
  zone   = "example.com"
  name   = "www"
  values = ["v=spf1 a -all"]
}
