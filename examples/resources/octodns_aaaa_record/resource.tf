resource "octodns_aaaa_record" "aaaa" {
  zone   = "example.com"
  name   = "aaaa"
  values = ["2601:644:500:e210:62f8:1dff:feb8:947a"]
}


