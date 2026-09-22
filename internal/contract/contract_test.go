package contract

import (
	"strings"
	"testing"
)

func TestLoadAndValidate(t *testing.T) {
	definition := `apiVersion: platform.example.io/v1alpha1
kind: Service
metadata:
  name: payment-api
spec:
  owner: payments-team
  image: ghcr.io/example/payment-api:v1.2.3
  port: 8080
  resources:
    size: small
  availability:
    replicas: 2
  observability:
    enabled: true
`
	service, err := Load(strings.NewReader(definition))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if errors := Validate(service); len(errors) != 0 {
		t.Fatalf("Validate() errors = %v", errors)
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	_, err := Load(strings.NewReader("apiVersion: platform.example.io/v1alpha1\nunknown: true\n"))
	if err == nil {
		t.Fatal("Load() accepted an unknown field")
	}
}

func TestValidateRejectsUnsafeValues(t *testing.T) {
	service := Service{
		APIVersion: APIVersion,
		Kind:       Kind,
		Metadata:   Metadata{Name: "Payment_API"},
		Spec: Spec{
			Owner:        "unknown-team",
			Image:        "example/payment-api:latest",
			Port:         70000,
			Resources:    Resources{Size: "huge"},
			Availability: Availability{Replicas: 1},
		},
	}
	if errors := Validate(service); len(errors) != 6 {
		t.Fatalf("Validate() returned %d errors, want 6: %v", len(errors), errors)
	}
}
