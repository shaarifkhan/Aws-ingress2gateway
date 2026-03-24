# Aws-ingress2gateway

This directory is a small standalone learning project for building an AWS-focused
Ingress to Gateway converter step by step.

## Current step

Step 1 through step 7:

- load Kubernetes client configuration from the local environment
- create a controller-runtime client
- list raw `Ingress` resources from the cluster
- filter the raw list down to ALB-backed `Ingress` objects
- load the filtered ALB ingresses into provider storage
- convert stored ALB ingresses into a tiny intermediate representation
- render a small human-readable summary of the generated model

Current IR shape:

- `Model`
- `Gateway`
- `HTTPRoute`
- link back to the source `Ingress`

Not implemented yet:

- real Gateway API objects
- YAML output
- richer field mapping from `Ingress` rules into routes and listeners
