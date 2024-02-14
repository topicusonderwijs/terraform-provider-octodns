locals {

  sshfp_raw = [
    { algorithm = 1, fingerprint = "bf6b6825d2977c511a475bbefb88aad54a92ac73", fingerprint_type = 1 },
    { algorithm = 1, fingerprint = "7491973e5f8b39d5327cd4e08bc81b05f7710b49", fingerprint_type = 1 }
  ]

  sshfp_values = [for v in local.sshfp_raw : "${v.algorithm} ${v.fingerprint_type} ${v.fingerprint}"]

}

resource "octodns_sshfp_record" "root" {
  zone   = "example.com"
  name   = "@"
  ttl    = 300
  values = local.sshfp_values

}

