package main

import (
	"fmt"
	"os"

	"github.com/user/envault/internal/cli"
	"github.com/user/envault/internal/store"
)

const usage = `envault — lightweight .env secret manager

Usage:
  envault set <KEY> <VALUE>   Store a secret
  envault get <KEY>           Retrieve a secret
  envault delete <KEY>        Remove a secret
  envault list                List all secret keys
  envault export [format]     Print shell exports (bash|fish|dotenv)
  envault import <file>       Import secrets from a .env file
`

func main() {
	vaultPath := store.VaultPath()

	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "set":
		if len(os.Args) != 4 {
			fatalf("usage: envault set <KEY> <VALUE>")
		}
		err = cli.RunSet(vaultPath, os.Args[2], os.Args[3])
	case "get":
		if len(os.Args) != 3 {
			fatalf("usage: envault get <KEY>")
		}
		err = cli.RunGet(vaultPath, os.Args[2])
	case "delete":
		if len(os.Args) != 3 {
			fatalf("usage: envault delete <KEY>")
		}
		err = cli.RunDelete(vaultPath, os.Args[2])
	case "list":
		err = cli.RunList(vaultPath)
	case "export":
		format := "bash"
		if len(os.Args) == 3 {
			format = os.Args[2]
		}
		err = cli.RunExport(vaultPath, format)
	case "import":
		if len(os.Args) != 3 {
			fatalf("usage: envault import <file>")
		}
		err = cli.RunImport(vaultPath, os.Args[2])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func fatalf(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
