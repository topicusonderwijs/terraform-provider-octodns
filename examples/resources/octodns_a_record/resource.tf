resource "octodns_a_record" "root" {
  zone   = "example.com"
  name   = "@"
  ttl    = 300
  values = ["1.2.3.4", "5.6.7.8"]
}


