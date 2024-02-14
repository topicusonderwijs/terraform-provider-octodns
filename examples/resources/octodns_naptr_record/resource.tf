locals {

  naptr_raw = [{
    flags       = "U"
    order       = 100
    preference  = 100
    regexp      = "'!^.*$!sip:info@bar.example.com!'"
    replacement = "."
    service     = "SIP+D2U"
  }]

  naptr_values = [for v in local.naptr_raw : "${v.order} ${v.preference} ${v.flags} ${v.service} ${v.regexp} ${v.replacement}"]

}

resource "octodns_naptr_record" "naptr" {
  zone   = "example.com"
  name   = "naptr"
  values = local.naptr_values
}



