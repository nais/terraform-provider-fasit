# Testing the provider locally

**1. Build the provider:**

```bash
mise run build
```

**2. Configure dev_overrides:**

Create a `dev.tfrc` file in the repo root (already gitignored):

```hcl
provider_installation {
  dev_overrides {
    "tfregistry.cloud.nais.io/nais/fasit" = "../path/to/terraform-provider-fasit"
  }
  direct {}
}
```

Then point OpenTofu at this file by setting `TF_CLI_CONFIG_FILE`:

```bash
export TF_CLI_CONFIG_FILE=$(pwd)/dev.tfrc
```

**3. Write a test config:**

```terraform
provider "fasit" {
  url      = "localhost:4444"
  insecure = true
}

resource "fasit_environment_value" "test" {
  environment_id = "<your-environment-id>"
  key            = "hello"
  value          = "world"
  secret         = false
}
```

**4. Port-forward to the Fasit service:**

```bash
kubectl port-forward svc/fasit 4444:4444 -n nais-system
```

**5. Run plan/apply:**

```bash
tofu plan
tofu apply
```
