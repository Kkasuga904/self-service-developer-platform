package command

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/example/self-service-developer-platform/internal/contract"
	"gopkg.in/yaml.v3"
)

func CreateService(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("create-service", flag.ContinueOnError)
	flags.SetOutput(stderr)
	name := flags.String("name", "", "service name")
	owner := flags.String("owner", "", "approved owning team")
	environment := flags.String("environment", "dev", "deployment environment: dev, staging, or prod")
	contact := flags.String("contact", "", "ownership identifier (team channel, email, or handle)")
	image := flags.String("image", "", "container image with a non-latest tag or digest")
	port := flags.Int("port", 0, "container port")
	size := flags.String("size", "small", "resource size: small, medium, or large")
	replicas := flags.Int("replicas", 2, "desired replicas, from 2 to 10")
	output := flags.String("output", "", "output path; defaults to services/OWNER/NAME/service.yaml")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}

	service := contract.Service{
		APIVersion: contract.APIVersion,
		Kind:       contract.Kind,
		Metadata:   contract.Metadata{Name: *name},
		Spec: contract.Spec{
			Owner:         *owner,
			Environment:   *environment,
			Contact:       *contact,
			Image:         *image,
			Port:          *port,
			Resources:     contract.Resources{Size: *size},
			Availability:  contract.Availability{Replicas: *replicas},
			Observability: contract.Observability{Enabled: true},
		},
	}
	if validationErrors := contract.Validate(service); len(validationErrors) > 0 {
		return joinValidationErrors(validationErrors)
	}

	path := *output
	if path == "" {
		path = filepath.Join("services", *owner, *name, "service.yaml")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("refusing to overwrite %s", path)
		}
		return fmt.Errorf("create %s: %w", path, err)
	}
	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	encodeErr := encoder.Encode(service)
	closeErr := file.Close()
	if encodeErr != nil {
		return fmt.Errorf("write %s: %w", path, encodeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close %s: %w", path, closeErr)
	}
	fmt.Fprintf(stdout, "created %s\n", path)
	return nil
}

func joinValidationErrors(validationErrors []error) error {
	return fmt.Errorf("service definition is invalid: %w", errors.Join(validationErrors...))
}
