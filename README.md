# Aws-ingress2gateway

`Aws-ingress2gateway` is a small standalone tool for converting ALB-backed
Kubernetes `Ingress` resources into Gateway API resources plus AWS-specific
Gateway customization manifests.

At the time of writing, the official SIG Network `ingress2gateway` project does
not support `aws-load-balancer-controller` / ALB ingress conversion into
Gateway API output. This project is designed to explore and fill that AWS ALB
conversion gap.

More detailed implementation notes, the 23-step buildout, and the current IR
shape live in `Implementation-details.md`.

## What it does

- reads `Ingress` resources from the current Kubernetes context
- filters them down to ALB-backed ingresses
- converts them into typed Gateway API `Gateway` and `HTTPRoute` resources
- generates AWS customization CRDs for load balancer and target group settings
- renders the generated resources as multi-document YAML

## Requirements

- Go `1.25.5` or later
- access to a Kubernetes cluster through your local kubeconfig
- ALB-backed `Ingress` resources managed by `aws-load-balancer-controller`

## Install

1. Clone this repository and change into `Aws-ingress2gateway`.
2. Download dependencies:

```bash
go mod download
```

3. Optionally build the CLI:

```bash
mkdir -p bin
go build -o bin/print-gateway-api-yaml ./cmd/print-gateway-api-yaml
```

## Run

Run directly with Go:

```bash
GOMODCACHE="$(pwd)/.gomodcache" GOSUMDB=off go run ./cmd/print-gateway-api-yaml --namespace default
```

Or run the built binary:

```bash
./bin/print-gateway-api-yaml --namespace default
```

## Flags

- `--namespace default` converts all ALB ingresses in one namespace
- `--ingress-name demo` converts the named ALB ingress across all namespaces
- `--namespace default --ingress-name demo` converts one named ALB ingress in one namespace
- no flags converts all ALB ingresses

## Output

The CLI prints combined YAML for:

- Gateway API `Gateway`
- Gateway API `HTTPRoute`
- AWS load balancer customization resources
- AWS target group customization resources
