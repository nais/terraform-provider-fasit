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

resource "fasit_environment_value" "name" {
  environment_id = fasit_environment.name.id
  key            = "key"
  value          = "value"
}
