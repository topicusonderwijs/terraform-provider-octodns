locals {

  srv_raw = [
    {
      service  = "imap"
      protocol = "tcp"
      port     = 30
      priority = 12
      target   = "foo-2.unit.tests."
      weight   = 20
      }, {
      service  = "imap"
      protocol = "tcp"
      port     = 30
      priority = 10
      target   = "foo-1.unit.tests."
      weight   = 20
    }
  ]

  srv_values = { for v in local.srv_raw : "_${v.service}._${v.protocol}" => "${v.priority} ${v.weight} ${v.port} ${v.target}"... }

}

resource "octodns_srv_record" "root" {
  for_each = local.srv_values
  zone     = "example.com"
  name     = each.key
  ttl      = 600
  values   = each.value
}


