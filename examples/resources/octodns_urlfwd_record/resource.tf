locals {

  urlfwd_raw = [{
    code    = 302
    masking = 2
    path    = "/"
    query   = 0
    target  = "http://www.unit.tests"
  }]

  urlfwd_values = [for v in local.urlfwd_raw : "${v.code} ${v.masking} ${v.path} ${v.query} ${v.target}"]

}

resource "octodns_urlfwd_record" "urlfwd" {
  zone   = "example.com"
  name   = "urlfwd"
  values = local.urlfwd_values
}

