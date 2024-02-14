resource "octodns_mx_record" "mx" {
  zone   = "example.com"
  name   = "mx"
  values = ["40 smtp-1.unit.tests.", "20 smtp-2.unit.tests."]
}



