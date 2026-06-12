#!/usr/bin/env bash
#MISE description="Validate generated docs (tfplugindocs validate)"
set -euo pipefail

go tool github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs validate --tf-version 1.10.5