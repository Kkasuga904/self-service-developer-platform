package command

import (
	"flag"
	"fmt"
	"io"
	"os/exec"
)

var requiredTools = []string{"git", "go", "terraform", "helm", "kubectl"}

func Doctor(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flags.SetOutput(stderr)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("doctor accepts no positional arguments")
	}

	missing := false
	for _, tool := range requiredTools {
		path, err := exec.LookPath(tool)
		if err != nil {
			fmt.Fprintf(stderr, "MISSING %s\n", tool)
			missing = true
			continue
		}
		fmt.Fprintf(stdout, "FOUND   %-9s %s\n", tool, path)
	}
	if missing {
		return fmt.Errorf("one or more required tools are missing")
	}
	return nil
}
