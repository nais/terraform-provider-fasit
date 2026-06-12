#!/usr/bin/env bash
#MISE description="Fetch fasit.proto from nais/fasit and generate the gRPC client"
set -euo pipefail

FASIT_REF="${FASIT_REF:-main}"
PROTO_URL="https://raw.githubusercontent.com/nais/fasit/${FASIT_REF}/schema/protobuf/fasit.proto"

GO_PACKAGE="github.com/nais/terraform-provider-fasit/fasit/protogen"
OUT_DIR="fasit/protogen"

tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT

echo "Fetching fasit.proto from nais/fasit@${FASIT_REF}"
curl -fsSL -o "${tmp}/fasit.proto" "${PROTO_URL}"

protoc \
  --proto_path="${tmp}" \
  --go_out="${OUT_DIR}" \
  --go_opt=paths=source_relative \
  --go_opt="Mfasit.proto=${GO_PACKAGE}" \
  --go-grpc_out="${OUT_DIR}" \
  --go-grpc_opt=paths=source_relative \
  --go-grpc_opt="Mfasit.proto=${GO_PACKAGE}" \
  fasit.proto

echo "Generated gRPC client in ${OUT_DIR}"
