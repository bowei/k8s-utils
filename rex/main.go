package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/bowei/k8s-misc/rex/pkg"
)

func main() {
	outputFile := flag.String("output", "-", "output file. '-' will write to stdout")
	jsonOutput := flag.Bool("json", false, "output JSON instead of HTML")
	startType := flag.String("type", "k8s.io/api/core/v1", "initial type to display")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Generate Go API documentation.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <package-directories...>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("[flag] -output=%q", *outputFile)
	log.Printf("[flag] -json=%t", *jsonOutput)
	log.Printf("[flag] -type=%v", *startType)
	log.Printf("[flag] packages=%v", args)

	allTypes, err := pkg.ParsePackages(args)
	if err != nil {
		log.Fatalf("Error parsing packages: %v", err)
	}

	var writer io.Writer
	if *outputFile == "-" {
		writer = os.Stdout
	} else {
		f, err := os.Create(*outputFile)
		if err != nil {
			log.Fatalf("Error creating file: %v", err)
		}
		defer f.Close()
		writer = f
	}

	if *jsonOutput {
		if err := pkg.WriteJSON(allTypes, writer); err != nil {
			log.Fatalf("Error writing JSON: %v", err)
		}
		return
	}

	log.Printf("Found %d types.\n", len(allTypes))

	if err := pkg.GenerateHTML(allTypes, writer, *startType); err != nil {
		log.Fatalf("Error generating HTML: %v", err)
	}
}
