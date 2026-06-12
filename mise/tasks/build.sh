#!/usr/bin/env bash
#MISE description="Build the provider"
set -euo pipefail

go build -o ./bin/terraform-provider-fasit .