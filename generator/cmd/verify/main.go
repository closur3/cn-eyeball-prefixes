package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/ipv4verify"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/ipv6build"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/listmanifest"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/publicverify"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/releasecheck"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	command := os.Args[1]
	os.Args = append([]string{os.Args[0]}, os.Args[2:]...)

	switch command {
	case "ipv4":
		ipv4verify.Main()
	case "ipv6":
		ipv6build.VerifyMain()
	case "manifest":
		flags := flag.NewFlagSet("verify manifest", flag.ExitOnError)
		root := flags.String("root", "", "public lists root")
		requirePublicationProvenance := flags.Bool(
			"require-publication-provenance",
			false,
			"require complete, canonical publication provenance",
		)
		currentRoot := flags.String(
			"current-root",
			"",
			"current public lists root for publication provenance comparison",
		)
		expectedNewGeneratorCommit := flags.String(
			"expected-new-generator-commit",
			"",
			"expected generator commit when candidate content changes",
		)
		if err := flags.Parse(os.Args[1:]); err != nil {
			panic(err)
		}
		if *root == "" {
			panic("--root is required")
		}
		if err := listmanifest.VerifyWithOptions(*root, listmanifest.VerifyOptions{
			RequirePublicationProvenance: *requirePublicationProvenance,
			CurrentRoot:                  *currentRoot,
			ExpectedNewGeneratorCommit:   *expectedNewGeneratorCommit,
		}); err != nil {
			panic(err)
		}
		fmt.Println("OK: public list manifest matches the complete list tree")
	case "public":
		flags := flag.NewFlagSet("verify public", flag.ExitOnError)
		root := flags.String("root", "", "public lists root")
		if err := flags.Parse(os.Args[1:]); err != nil {
			panic(err)
		}
		if *root == "" {
			panic("--root is required")
		}
		if err := publicverify.Verify(*root); err != nil {
			panic(err)
		}
		fmt.Println("OK: public list operator and province relationships are valid")
	case "release":
		releasecheck.Main()
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: verify <ipv4|ipv6|manifest|public|release> [flags]")
	os.Exit(2)
}
