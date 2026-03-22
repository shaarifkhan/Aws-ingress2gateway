# Aws-ingress2gateway

This directory is a small standalone learning project for building an AWS-focused
Ingress to Gateway converter step by step.

## Current step

Step 1 through step 4:

- load Kubernetes client configuration from the local environment
- create a controller-runtime client
- list raw `Ingress` resources from the cluster
- filter the raw list down to ALB-backed `Ingress` objects
- load the filtered ALB ingresses into provider storage

Not implemented yet:

- intermediate representation
- Gateway API conversion
- YAML output
