# To test the provider locally

Create a directory for your architecture as follows:

```bash
mkdir -p ~/.terraform.d/plugins/registry.local/nais/fasit/1.0.0-local/<arch>/
```

Linux AMD64:

```bash
mkdir -p ~/.terraform.d/plugins/registry.local/nais/fasit/1.0.0-local/linux_amd64/
```

Build the plugin into the directory:

```bash
go build -o ~/.terraform.d/plugins/registry.local/nais/fasit/1.0.0-local/<arch>/terraform-provider-fasit_v1.0.0-local

# Example for Linux AMD64
go build -o ~/.terraform.d/plugins/registry.local/nais/fasit/1.0.0-local/linux_amd64/terraform-provider-fasit_v1.0.0-local
```

Add the following to your terraform file:

```terraform
terraform {
  required_providers {
    fasit = {
      source  = "registry.local/nais/fasit"
      version = "1.0.0-local"
    }
  }
}

provider "fasit" {
  url = "http://localhost:4444"
  insecure = true
}
```

Run `terraform init` to initialize the plugin.
