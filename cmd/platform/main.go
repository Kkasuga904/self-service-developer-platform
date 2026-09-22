package main

import (
	"fmt"
	"os"

	"github.com/example/self-service-developer-platform/internal/command"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "create-service":
		err = command.CreateService(os.Args[2:], os.Stdout, os.Stderr)
	case "validate":
		err = command.Validate(os.Args[2:], os.Stdout, os.Stderr)
	case "doctor":
		err = command.Doctor(os.Args[2:], os.Stdout, os.Stderr)
	case "help", "-h", "--help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `platform manages the Platform Service contract.

Usage:
  platform create-service --name NAME --owner OWNER --image IMAGE --port PORT [--output PATH]
  platform validate PATH [PATH...]
  platform doctor

Commands:
  create-service  Create a review-ready Service Definition without overwriting files
  validate        Strictly parse and validate one or more Service Definitions
  doctor          Check local tools used by the developer workflow`)
}
