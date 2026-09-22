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
  environment: dev
  contact: payments-team@example.com
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
			Environment:  "qa",
			Contact:      "   ",
			Image:        "example/payment-api:latest",
			Port:         70000,
			Resources:    Resources{Size: "huge"},
			Availability: Availability{Replicas: 1},
		},
	}
	if errors := Validate(service); len(errors) != 8 {
		t.Fatalf("Validate() returned %d errors, want 8: %v", len(errors), errors)
	}
}

func TestValidateAcceptsSecondTeam(t *testing.T) {
	service := Service{
		APIVersion: APIVersion,
		Kind:       Kind,
		Metadata:   Metadata{Name: "order-api"},
		Spec: Spec{
			Owner:         "orders-team",
			Environment:   "dev",
			Image:         "ghcr.io/example/order-api:v0.2.0",
			Port:          8080,
			Resources:     Resources{Size: "small"},
			Availability:  Availability{Replicas: 2},
			Observability: Observability{Enabled: true},
		},
	}
	if errors := Validate(service); len(errors) != 0 {
		t.Fatalf("Validate() errors = %v", errors)
	}
	if namespace, ok := NamespaceForOwner("orders-team"); !ok || namespace != "team-orders" {
		t.Fatalf("NamespaceForOwner(orders-team) = %q, %v", namespace, ok)
	}
}
