resource "octodns_alias_record" "root" {
  zone   = "example.com"
  name   = "@"
  ttl    = 300
  values = ["www.example.com."]
}
