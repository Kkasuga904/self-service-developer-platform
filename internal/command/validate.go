package command

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/example/self-service-developer-platform/internal/contract"
)

func Validate(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("validate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() == 0 {
		return fmt.Errorf("at least one Service Definition path is required")
	}

	failed := false
	for _, path := range flags.Args() {
		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(stderr, "FAIL %s: %v\n", path, err)
			failed = true
			continue
		}
		service, loadErr := contract.Load(file)
		closeErr := file.Close()
		if loadErr != nil {
			fmt.Fprintf(stderr, "FAIL %s: %v\n", path, loadErr)
			failed = true
			continue
		}
		if closeErr != nil {
			fmt.Fprintf(stderr, "FAIL %s: %v\n", path, closeErr)
			failed = true
			continue
		}
		validationErrors := contract.Validate(service)
		if len(validationErrors) > 0 {
			for _, validationErr := range validationErrors {
				fmt.Fprintf(stderr, "FAIL %s: %v\n", path, validationErr)
			}
			failed = true
			continue
		}
		fmt.Fprintf(stdout, "PASS %s\n", path)
	}
	if failed {
		return fmt.Errorf("one or more Service Definitions are invalid")
	}
	return nil
}
