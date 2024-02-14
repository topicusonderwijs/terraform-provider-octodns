resource "octodns_txt_record" "root" {
  zone = "example.com"
  name = "@"
  ttl  = 300
  values = [
    "Bah bah black sheep",
    "have you any wool.",
    "v=DKIM1\\;k=rsa\\;s=email\\;h=sha256\\;p=A/kinda+of/long/string+with+numb3rs",
  ]
}


