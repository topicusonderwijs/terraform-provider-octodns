provider "octodns" {
  github_access_token = "ghp_xxxxxxxxxxxxx"
  github_org          = "example_org"
  github_repo         = "dns_repo"

  scope {
    name = "default"
    path = "zones"
  }

}

