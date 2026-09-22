package command

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateServiceAndValidate(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "service.yaml")
	var stdout, stderr bytes.Buffer
	err := CreateService([]string{
		"--name", "payment-api",
		"--owner", "payments-team",
		"--image", "ghcr.io/example/payment-api:v1.2.3",
		"--port", "8080",
		"--output", path,
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("CreateService() error = %v, stderr = %s", err, stderr.String())
	}
	if err := Validate([]string{path}, &stdout, &stderr); err != nil {
		t.Fatalf("Validate() error = %v, stderr = %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "PASS") {
		t.Fatalf("output %q does not report PASS", stdout.String())
	}
}

func TestCreateServiceDoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.yaml")
	if err := os.WriteFile(path, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := CreateService([]string{
		"--name", "payment-api", "--owner", "payments-team",
		"--image", "example/payment-api:v1", "--port", "8080", "--output", path,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("CreateService() error = %v, want overwrite refusal", err)
	}
}

func TestValidateRejectsLatest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.yaml")
	definition := strings.Replace(validDefinition, ":v1.2.3", ":latest", 1)
	if err := os.WriteFile(path, []byte(definition), 0o644); err != nil {
		t.Fatal(err)
	}
	err := Validate([]string{path}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("Validate() accepted latest tag")
	}
}

const validDefinition = `apiVersion: platform.example.io/v1alpha1
kind: Service
metadata:
  name: payment-api
spec:
  owner: payments-team
  image: example/payment-api:v1.2.3
  port: 8080
  resources:
    size: small
  availability:
    replicas: 2
  observability:
    enabled: true
`
