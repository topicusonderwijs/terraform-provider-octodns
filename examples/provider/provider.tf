provider "octodns" {
  github_access_token = "ghp_xxxxxxxxxxxxx"
  github_org          = "example_org"
  github_repo         = "dns_repo"

  scope {
    path = "zones"
  }

}


provider "octodns" {
  github_access_token = "ghp_xxxxxxxxxxxxx"
  github_org          = "example_org"
  github_repo         = "dns_repo"

  scope {
    path = "internal/zones"
  }
  scope {
    name = "external"
    path = "external/zones"
  }

}
