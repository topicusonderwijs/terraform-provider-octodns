data "octodns_a_record" "example" {
  zone  = "example.com"
  scope = "default"
  name  = "localhost"
}
