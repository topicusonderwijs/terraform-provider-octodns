data "octodns_a_record" "default_scope" {
  zone = "example.com"
  name = "@"
}


data "octodns_a_record" "custom_scope" {
  zone  = "example.com"
  scope = "internal"
  name  = "www"
}
