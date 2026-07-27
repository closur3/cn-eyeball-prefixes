package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/closur3/cn-eyeball-prefixes/generator/internal/ipv4build"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/ipv6build"
	"github.com/closur3/cn-eyeball-prefixes/generator/internal/listmanifest"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	command := os.Args[1]
	os.Args = append([]string{os.Args[0]}, os.Args[2:]...)

	switch command {
	case "ipv4":
		ipv4build.Main()
	case "ipv6":
		ipv6build.Main()
	case "manifest":
		flags := flag.NewFlagSet("manifest", flag.ExitOnError)
		root := flags.String("root", "", "public lists root")
		commit := flags.String("commit", "", "generator git commit hash")
		dirty := flags.Bool("dirty", false, "generator working tree has uncommitted changes")
		configHashes := flags.String("config-hashes", "", "path to JSON file with config file hashes (name -> {sha256})")
		sourceDir := flags.String("source-dir", "", "path to upstream source directory (reads all files, computes sha256, looks up URLs)")
		if err := flags.Parse(os.Args[1:]); err != nil {
			panic(err)
		}
		if *root == "" {
			panic("--root is required")
		}

		var gen *listmanifest.GeneratorInfo
		if *commit != "" {
			gen = &listmanifest.GeneratorInfo{Commit: *commit, Dirty: *dirty}
		}

		var configs map[string]listmanifest.SourceEntry
		if *configHashes != "" {
			b, err := os.ReadFile(*configHashes)
			if err != nil {
				panic(err)
			}
			if err := json.Unmarshal(b, &configs); err != nil {
				panic(err)
			}
		}

		var sources map[string]listmanifest.SourceEntry
		if *sourceDir != "" {
			var err error
			sources, err = listmanifest.ComputeSourceHashes(*sourceDir)
			if err != nil {
				panic(err)
			}
		}

		changed, err := listmanifest.Generate(*root, time.Now(), gen, configs, sources)
		if err != nil {
			panic(err)
		}
		if changed {
			fmt.Println("updated public list manifest")
		} else {
			fmt.Println("public list manifest is already current")
		}
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: generate <ipv4|ipv6|manifest> [flags]")
	os.Exit(2)
}
