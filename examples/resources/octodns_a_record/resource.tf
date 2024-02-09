resource "octodns_a_record" "localhost" {
  zone   = "example.com"
  name   = "localhost"
  ttl    = 3600
  values = ["127.0.0.1"]
  octodns = {
    cloudflare = {
      proxied = true
    }
  }
}

resource "octodns_a_record" "minimal" {
  zone   = "example.com"
  name   = "www"
  values = ["127.0.0.1"]
}
