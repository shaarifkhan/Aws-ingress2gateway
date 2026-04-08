# Aws-ingress2gateway

This directory is a small standalone learning project for building an AWS-focused
Ingress to Gateway converter step by step.

## Current step

Step 1 through step 23:

- load Kubernetes client configuration from the local environment
- create a controller-runtime client
- list raw `Ingress` resources from the cluster
- filter the raw list down to ALB-backed `Ingress` objects
- load the filtered ALB ingresses into provider storage
- convert stored ALB ingresses into a tiny intermediate representation
- render a small human-readable summary of the generated model
- copy basic host, path, and backend service data from `Ingress` rules into the IR
- derive tiny gateway listeners from ingress hostnames
- link each generated route back to its generated gateway listeners
- distinguish basic HTTP and HTTPS listeners from ingress TLS configuration
- read ALB `listen-ports` annotation to drive generated listener ports
- read ALB `scheme` annotation into the gateway IR
- read ALB `target-type` annotation into backend refs
- convert the tiny IR into real typed Gateway API `Gateway` and `HTTPRoute` objects
- render typed Gateway API objects as multi-document YAML
- expose typed Gateway API resources and YAML from the provider layer
- improve typed manifest fidelity by splitting routes per hostname and carrying AWS annotations forward
- add a tiny CLI that prints provider-generated Gateway API YAML
- add initial `LoadBalancerConfiguration` IR for LB-focused ingress annotations
- add initial `TargetGroupConfiguration` IR for target type and health check path
- convert load balancer and target group IR into typed AWS Gateway customization CRDs
- map ALB load balancer attributes and WAFv2 ACL into load balancer IR and typed AWS CRDs

Current IR shape:

- `Model`
- `Gateway`
- `Listener`
- `HTTPRoute`
- `ParentRef`
- `HTTPRouteRule`
- `BackendRef`
- `LoadBalancerConfiguration`
- `LoadBalancerListenerConfiguration`
- `TargetGroupConfiguration`
- link back to the source `Ingress`
- basic HTTP listener data on the gateway side
- basic HTTPS listener data when ingress TLS is present
- ALB listener port overrides from `alb.ingress.kubernetes.io/listen-ports`
- ALB scheme capture from `alb.ingress.kubernetes.io/scheme`
- ALB target type capture from `alb.ingress.kubernetes.io/target-type`
- route parent references to generated gateway listener sections
- hostname, path, and backend service details from `Ingress.Spec.Rules`
- typed Gateway API objects built from the current IR
- multi-document YAML output from the typed Gateway API objects
- provider methods for end-to-end Gateway API resources and YAML
- hostname-aware `HTTPRoute` splitting in typed Gateway API output
- a tiny demo CLI that prints generated Gateway API YAML plus AWS customization CRD YAML
- initial load balancer configuration IR for name, scheme, and listener configurations like certs and SSL policy
- load balancer attribute capture from `alb.ingress.kubernetes.io/load-balancer-attributes`
- WAFv2 ACL capture from `alb.ingress.kubernetes.io/wafv2-acl-arn`
- initial target group configuration IR for service-level target type and health check path
- typed AWS Gateway customization CRDs built from the current load balancer and target group IR
- generated `Gateway.spec.infrastructure.parametersRef` pointing at typed `LoadBalancerConfiguration`

Current demo command:

```bash
GOMODCACHE="$(pwd)/.gomodcache" GOSUMDB=off go run ./cmd/print-gateway-api-yaml --namespace default
```

Optional flags:

- `--namespace default` converts all ALB ingresses in one namespace
- `--ingress-name demo` converts the named ALB ingress across all namespaces
- `--namespace default --ingress-name demo` converts one named ALB ingress in one namespace
- no flags converts all ALB ingresses

Not implemented yet:

- default backend handling and more advanced ALB-specific annotations
- richer manifest fidelity beyond hostname splitting and basic typed field mapping
