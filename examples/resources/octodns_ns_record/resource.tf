resource "octodns_ns_record" "root" {
  zone   = "example.com"
  name   = "@"
  ttl    = 300
  values = ["2.2.2.2", "3.3.3.3"]
}


