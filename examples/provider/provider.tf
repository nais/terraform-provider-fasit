terraform {
  required_providers {
    fasit = {
      source = "tfregistry.cloud.nais.io/nais/fasit"
    }
  }
}

provider "fasit" {
  insecure = true
  url      = "http://localhost:8080"
  # example configuration here
}

resource "fasit_tenant" "name" {
  name = "test-tenant"
}


resource "fasit_environment" "name" {
  tenant_id = fasit_tenant.name.id
  name      = "test"
}

resource "fasit_environment_value" "example" {
  environment_id = fasit_environment.name.id
  key            = "MY_SECRET"
  value          = "my-value"
  hide_in_fasit  = true
}
