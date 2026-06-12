#!/usr/bin/env bash
#MISE description="Update all Go tools to latest"
set -euo pipefail

# Collect tools from the `tool (...)` block in go.mod and update them to latest.
tools=()
for tool in $(go list tool); do
  tools+=("${tool}@latest")
done

if [ "${#tools[@]}" -eq 0 ]; then
  echo "No tools found in go.mod" >&2
  exit 1
fi

go get -tool "${tools[@]}"
go mod tidy
