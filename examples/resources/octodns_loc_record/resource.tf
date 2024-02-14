locals {

  loc_raw = [
    {
      altitude       = 20
      lat_degrees    = 31
      lat_direction  = "S"
      lat_minutes    = 58
      lat_seconds    = 52.1
      long_degrees   = 115
      long_direction = "E"
      long_minutes   = 49
      long_seconds   = 11.7
      precision_horz = 10
      precision_vert = 2
      size           = 10
    },
    {
      altitude       = 20
      lat_degrees    = 53
      lat_direction  = "N"
      lat_minutes    = 13
      lat_seconds    = 10
      long_degrees   = 2
      long_direction = "W"
      long_minutes   = 18
      long_seconds   = 26
      precision_horz = 1000
      precision_vert = 2
      size           = 10
    }
  ]

  loc_values = [for v in local.loc_raw :
  "${v.lat_degrees} ${v.lat_minutes} ${v.lat_seconds} ${v.lat_direction} ${v.long_degrees} ${v.long_minutes} ${v.long_seconds} ${v.long_direction} ${v.altitude} ${v.size} ${v.precision_horz} ${v.precision_vert}"]

}

resource "octodns_loc_record" "loc" {
  zone   = "example.com"
  name   = "loc"
  values = local.loc_values
}


