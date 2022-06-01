FROM golang:1.18-alpine as builder

RUN apk add --no-cache git

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . /workspace

# Build
RUN CGO_ENABLED=0 go build -a -o terraform-provider-fasit

FROM ghcr.io/runatlantis/atlantis:v0.19.4-pre.20220513

COPY --from=builder /workspace/terraform-provider-fasit /usr/local/bin/