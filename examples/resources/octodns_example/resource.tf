resource "octodns_a_record" "example" {

  zone  = "example.com"
  scope = "default"
  name  = "localhost"

  ttl    = "3600"
  values = ["127.0.0.1"]

}


