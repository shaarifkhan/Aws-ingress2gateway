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
	ingressName := flag.String("ingress-name", "", "specific ingress to convert; empty means all ingresses")
	flag.Parse()

	provider := albprovider.NewProvider()
	if err := provider.LoadFromCluster(context.Background(), *namespace); err != nil {
		fmt.Fprintf(os.Stderr, "load ingresses: %v\n", err)
		os.Exit(1)
	}

	provider.FilterStoredIngresses(*namespace, *ingressName)

	rendered, err := provider.BuildCombinedYAML()
	if err != nil {
		fmt.Fprintf(os.Stderr, "render combined yaml: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(rendered)
}
