# Aws-ingress2gateway

This directory is a small standalone learning project for building an AWS-focused
Ingress to Gateway converter step by step.

## Current step

Step 1 through step 15:

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

Current IR shape:

- `Model`
- `Gateway`
- `Listener`
- `HTTPRoute`
- `ParentRef`
- `HTTPRouteRule`
- `BackendRef`
- link back to the source `Ingress`
- basic HTTP listener data on the gateway side
- basic HTTPS listener data when ingress TLS is present
- ALB listener port overrides from `alb.ingress.kubernetes.io/listen-ports`
- ALB scheme capture from `alb.ingress.kubernetes.io/scheme`
- ALB target type capture from `alb.ingress.kubernetes.io/target-type`
- route parent references to generated gateway listener sections
- hostname, path, and backend service details from `Ingress.Spec.Rules`
- typed Gateway API objects built from the current IR

Not implemented yet:

- YAML output
- richer listener and route mapping behavior beyond basic HTTP/HTTPS, scheme, port, and target-type extraction
- default backend handling and more advanced ALB-specific annotations
