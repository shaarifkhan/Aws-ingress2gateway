package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"aws-ingress2gateway/pkg/albprovider"
)

func main() {
	namespace := flag.String("namespace", "", "namespace to read ingresses from; empty means all namespaces")
	flag.Parse()

	provider := albprovider.NewProvider()
	if err := provider.LoadFromCluster(context.Background(), *namespace); err != nil {
		fmt.Fprintf(os.Stderr, "load ingresses: %v\n", err)
		os.Exit(1)
	}

	rendered, err := provider.BuildGatewayAPIYAML()
	if err != nil {
		fmt.Fprintf(os.Stderr, "render gateway api yaml: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(rendered)
}
