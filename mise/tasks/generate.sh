#!/usr/bin/env bash
#MISE description="Generate code/docs (go generate)"
set -euo pipefail

go generate ./...
