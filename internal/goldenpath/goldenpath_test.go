// Package goldenpath renders the Golden Path chart for every committed
// Service Definition with the real helm binary and asserts the platform
// invariants: ownership metadata, secure defaults, probes, resources and
// pinned images. This is the Case E ("valid Golden Path workload") check
// against the actual artifacts, complementing the Kyverno Pod fixtures.
package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/example/self-service-developer-platform/internal/contract"
	"gopkg.in/yaml.v3"
)

type service struct {
	definition contract.Service
	path       string
	release    string
}

func committedServices(t *testing.T) []service {
	t.Helper()
	root := filepath.Join("..", "..")
	definitions := []struct {
		path    string
		release string
	}{
		{filepath.Join("services", "payments-team", "payment-api", "service.yaml"), "payment-api"},
		{filepath.Join("services", "orders-team", "order-api", "service.yaml"), "order-api"},
	}
	var services []service
	for _, definition := range definitions {
		full := filepath.Join(root, definition.path)
		file, err := os.Open(full)
		if err != nil {
			t.Fatalf("open %s: %v", definition.path, err)
		}
		loaded, err := contract.Load(file)
		file.Close()
		if err != nil {
			t.Fatalf("Load(%s): %v", definition.path, err)
		}
		if errors := contract.Validate(loaded); len(errors) != 0 {
			t.Fatalf("Validate(%s): %v", definition.path, errors)
		}
		services = append(services, service{definition: loaded, path: full, release: definition.release})
	}
	return services
}

func render(t *testing.T, chart string, service service) []map[string]any {
	t.Helper()
	helm, err := exec.LookPath("helm")
	if err != nil {
		t.Skip("helm executable not found")
	}
	root := filepath.Join("..", "..")
	output, err := exec.Command(helm, "template", service.release,
		filepath.Join(root, chart), "-f", service.path).CombinedOutput()
	if err != nil {
		t.Fatalf("helm template %s: %v\n%s", service.release, err, output)
	}
	var documents []map[string]any
	for _, chunk := range strings.Split(string(output), "\n---") {
		var lines []string
		for _, line := range strings.Split(chunk, "\n") {
			if trimmed := strings.TrimSpace(line); trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			lines = append(lines, line)
		}
		if len(lines) == 0 {
			continue
		}
		var document map[string]any
		if err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &document); err != nil {
			t.Fatalf("parse rendered %s: %v", service.release, err)
		}
		if len(document) == 0 {
			continue
		}
		documents = append(documents, document)
	}
	return documents
}

func byKind(documents []map[string]any, kind string) []map[string]any {
	var matched []map[string]any
	for _, document := range documents {
		if document["kind"] == kind {
			matched = append(matched, document)
		}
	}
	return matched
}

func nested(document map[string]any, fields ...string) any {
	var current any = document
	for _, field := range fields {
		mapping, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = mapping[field]
	}
	return current
}

func nonEmpty(t *testing.T, context string, value any) string {
	t.Helper()
	text, ok := value.(string)
	if !ok || text == "" {
		t.Fatalf("%s must be a non-empty string, got %v", context, value)
	}
	return text
}

func TestBothTeamsRenderSecureWorkloads(t *testing.T) {
	services := committedServices(t)
	if len(services) != 2 {
		t.Fatalf("want 2 committed services, got %d", len(services))
	}
	for _, service := range services {
		service := service
		t.Run(service.release, func(t *testing.T) {
			documents := render(t, filepath.Join("charts", "golden-path"), service)

			deployments := byKind(documents, "Deployment")
			if len(deployments) != 1 {
				t.Fatalf("want 1 Deployment, got %d", len(deployments))
			}
			if len(byKind(documents, "Service")) != 1 || len(byKind(documents, "PodDisruptionBudget")) != 1 {
				t.Fatal("want 1 Service and 1 PodDisruptionBudget")
			}
			deployment := deployments[0]

			labels, _ := nested(deployment, "metadata", "labels").(map[string]any)
			if labels == nil {
				t.Fatal("Deployment has no labels")
			}
			if got := nonEmpty(t, "owner label", labels["platform.example.io/owner"]); got != service.definition.Spec.Owner {
				t.Errorf("owner label = %q, want %q", got, service.definition.Spec.Owner)
			}
			if got := nonEmpty(t, "service label", labels["platform.example.io/service"]); got != service.definition.Metadata.Name {
				t.Errorf("service label = %q, want %q", got, service.definition.Metadata.Name)
			}
			if got := nonEmpty(t, "environment label", labels["platform.example.io/environment"]); got != service.definition.Spec.Environment {
				t.Errorf("environment label = %q, want %q", got, service.definition.Spec.Environment)
			}
			nonEmpty(t, "managed-by label", labels["app.kubernetes.io/managed-by"])

			if replicas, _ := nested(deployment, "spec", "replicas").(int); replicas != service.definition.Spec.Availability.Replicas {
				t.Errorf("replicas = %v, want %d", nested(deployment, "spec", "replicas"), service.definition.Spec.Availability.Replicas)
			}

			podSpec, _ := nested(deployment, "spec", "template", "spec").(map[string]any)
			if podSpec == nil {
				t.Fatal("Deployment has no pod spec")
			}
			if runAsNonRoot, _ := nested(podSpec, "securityContext", "runAsNonRoot").(bool); !runAsNonRoot {
				t.Error("pod runAsNonRoot must be true")
			}
			containers, _ := nested(podSpec, "containers").([]any)
			if len(containers) != 1 {
				t.Fatalf("want 1 container, got %d", len(containers))
			}
			container, _ := containers[0].(map[string]any)
			security, _ := nested(container, "securityContext").(map[string]any)
			if security == nil {
				t.Fatal("container has no securityContext")
			}
			if allow, _ := security["allowPrivilegeEscalation"].(bool); allow {
				t.Error("allowPrivilegeEscalation must be false")
			}
			if privileged, _ := security["privileged"].(bool); privileged {
				t.Error("privileged must not be true")
			}
			drop, _ := nested(security, "capabilities", "drop").([]any)
			if len(drop) != 1 || drop[0] != "ALL" {
				t.Errorf("capabilities.drop = %v, want [ALL]", nested(security, "capabilities", "drop"))
			}
			if readOnly, _ := security["readOnlyRootFilesystem"].(bool); !readOnly {
				t.Error("readOnlyRootFilesystem must be true")
			}
			for _, field := range [][]string{
				{"resources", "requests", "cpu"}, {"resources", "requests", "memory"},
				{"resources", "limits", "cpu"}, {"resources", "limits", "memory"},
			} {
				nonEmpty(t, "container "+strings.Join(field, "."),
					nested(container, field...))
			}
			if nested(container, "readinessProbe", "httpGet", "path") == nil {
				t.Error("readinessProbe.httpGet.path is required")
			}
			if nested(container, "livenessProbe", "httpGet", "path") == nil {
				t.Error("livenessProbe.httpGet.path is required")
			}
			image, _ := container["image"].(string)
			if image == "" || strings.HasSuffix(image, ":latest") {
				t.Errorf("image %q must be pinned and not latest", image)
			} else if !strings.Contains(image, ":") && !strings.Contains(image, "@") {
				t.Errorf("image %q must carry a tag or digest", image)
			}
		})
	}
}
