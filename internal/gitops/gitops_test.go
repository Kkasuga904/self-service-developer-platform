// Package gitops asserts the multi-team GitOps structure without a cluster:
// every approved owner maps to a labeled namespace, the AppProject allows
// only those namespaces with a fixed resource whitelist, the ApplicationSet
// discovers every committed service, and team RBAC grants no cross-team
// writes. Developers never receive direct workload-write access because
// GitOps owns the desired state.
package gitops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/example/self-service-developer-platform/internal/contract"
	"gopkg.in/yaml.v3"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func loadYAML(t *testing.T, path string) map[string]any {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var document map[string]any
	if err := yaml.Unmarshal(content, &document); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return document
}

// loadMultiYAML loads a file that may contain several `---` separated documents.
func loadMultiYAML(t *testing.T, path string) []map[string]any {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var documents []map[string]any
	for _, chunk := range strings.Split(string(content), "\n---") {
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
			t.Fatalf("parse %s: %v", path, err)
		}
		documents = append(documents, document)
	}
	return documents
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

func committedOwners(t *testing.T) map[string]string {
	t.Helper()
	root := repoRoot(t)
	owners := map[string]string{}
	for _, owner := range contract.ApprovedOwners() {
		namespace, ok := contract.NamespaceForOwner(owner)
		if !ok {
			t.Fatalf("no namespace mapping for approved owner %q", owner)
		}
		owners[owner] = namespace
		pattern := filepath.Join(root, "services", owner, "*", "service.yaml")
		matches, err := filepath.Glob(pattern)
		if err != nil || len(matches) == 0 {
			t.Fatalf("owner %q has no committed service", owner)
		}
		for _, match := range matches {
			file, err := os.Open(match)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := contract.Load(file)
			file.Close()
			if err != nil {
				t.Fatalf("Load(%s): %v", match, err)
			}
			if loaded.Spec.Owner != owner {
				t.Errorf("%s owner = %q, want directory owner %q", match, loaded.Spec.Owner, owner)
			}
			if errors := contract.Validate(loaded); len(errors) != 0 {
				t.Fatalf("Validate(%s): %v", match, errors)
			}
		}
	}
	if len(owners) < 2 {
		t.Fatalf("want at least 2 teams, got %v", owners)
	}
	return owners
}

func TestTeamNamespacesLabeled(t *testing.T) {
	owners := committedOwners(t)
	root := repoRoot(t)
	namespaceFiles := map[string]string{
		"team-payments": filepath.Join(root, "gitops", "platform", "team-payments-namespace.yaml"),
		"team-orders":   filepath.Join(root, "gitops", "platform", "team-orders-namespace.yaml"),
	}
	seen := map[string]bool{}
	for owner, namespace := range owners {
		path, ok := namespaceFiles[namespace]
		if !ok {
			t.Fatalf("no namespace manifest for %q", namespace)
		}
		document := loadYAML(t, path)
		if document["kind"] != "Namespace" {
			t.Errorf("%s kind = %v, want Namespace", path, document["kind"])
		}
		if name, _ := nested(document, "metadata", "name").(string); name != namespace {
			t.Errorf("%s name = %q, want %q", path, name, namespace)
		}
		labels, _ := nested(document, "metadata", "labels").(map[string]any)
		if labels["platform.example.io/owner"] != owner {
			t.Errorf("%s owner label = %v, want %q", path, labels["platform.example.io/owner"], owner)
		}
		seen[namespace] = true
	}
	if len(seen) != len(namespaceFiles) {
		t.Errorf("namespace manifests and owners diverged: %v", seen)
	}
}

func TestAppProjectScope(t *testing.T) {
	owners := committedOwners(t)
	document := loadYAML(t, filepath.Join(repoRoot(t), "gitops", "platform", "project.yaml"))
	destinations, _ := nested(document, "spec", "destinations").([]any)
	allowed := map[string]bool{}
	for _, destination := range destinations {
		mapping, _ := destination.(map[string]any)
		namespace, _ := mapping["namespace"].(string)
		server, _ := mapping["server"].(string)
		if server != "https://kubernetes.default.svc" {
			t.Errorf("destination server = %q, want in-cluster", server)
		}
		allowed[namespace] = true
	}
	for _, namespace := range owners {
		if !allowed[namespace] {
			t.Errorf("AppProject missing destination for namespace %q", namespace)
		}
	}
	if len(allowed) != len(owners) {
		t.Errorf("AppProject destinations %v do not match team namespaces", allowed)
	}
	whitelist, _ := nested(document, "spec", "namespaceResourceWhitelist").([]any)
	kinds := map[string]bool{}
	for _, entry := range whitelist {
		mapping, _ := entry.(map[string]any)
		kinds[mapping["kind"].(string)] = true
	}
	for _, kind := range []string{"Deployment", "Service", "HorizontalPodAutoscaler", "PodDisruptionBudget"} {
		if !kinds[kind] {
			t.Errorf("AppProject whitelist missing %q", kind)
		}
	}
}

func TestApplicationSetDiscoversServices(t *testing.T) {
	owners := committedOwners(t)
	document := loadYAML(t, filepath.Join(repoRoot(t), "gitops", "platform", "applicationset.yaml"))
raw, _ := nested(document, "spec", "generators").([]any)
	if len(raw) == 0 {
		t.Fatal("ApplicationSet has no generators")
	}
	generator, _ := raw[0].(map[string]any)
	files, _ := nested(generator, "git", "files").([]any)
	matched := false
	for _, file := range files {
		mapping, _ := file.(map[string]any)
		if mapping["path"] == "services/*/*/service.yaml" {
			matched = true
		}
	}
	if !matched {
		t.Error("ApplicationSet does not glob services/*/*/service.yaml")
	}
	destination, _ := nested(document, "spec", "template", "spec", "destination", "namespace").(string)
	if !strings.Contains(destination, "owner") {
		t.Errorf("ApplicationSet destination %q is not derived from the service owner", destination)
	}
	for owner := range owners {
		_ = owner
	}
}

func TestRBACHasNoCrossTeamWrite(t *testing.T) {
	owners := committedOwners(t)
	root := repoRoot(t)
	rbacFiles := map[string]string{
		"team-payments": filepath.Join(root, "gitops", "platform", "team-payments-rbac.yaml"),
		"team-orders":   filepath.Join(root, "gitops", "platform", "team-orders-rbac.yaml"),
	}
	groups := map[string]string{}
	for owner, namespace := range owners {
		path, ok := rbacFiles[namespace]
		if !ok {
			t.Fatalf("no RBAC manifest for %q", namespace)
		}
		var role, binding map[string]any
		for _, document := range loadMultiYAML(t, path) {
			kind, _ := document["kind"].(string)
			metadataNamespace, _ := nested(document, "metadata", "namespace").(string)
			if metadataNamespace != namespace {
				t.Errorf("%s %s lives in %q, want %q", path, kind, metadataNamespace, namespace)
			}
			switch kind {
			case "Role":
				role = document
			case "RoleBinding":
				binding = document
			default:
				t.Errorf("%s has unexpected kind %q (only Role/RoleBinding allowed)", path, kind)
			}
		}
		if role == nil || binding == nil {
			t.Fatalf("%s must contain exactly one Role and one RoleBinding", path)
		}
		rules, _ := role["rules"].([]any)
		if len(rules) == 0 {
			t.Fatalf("%s Role has no rules", path)
		}
		for _, item := range rules {
			rule, _ := item.(map[string]any)
			resources, _ := rule["resources"].([]any)
			verbs, _ := rule["verbs"].([]any)
			verbSet := map[string]bool{}
			for _, verb := range verbs {
				verbSet[verb.(string)] = true
			}
			for _, forbidden := range []string{"create", "update", "patch", "delete", "deletecollection", "*"} {
				if verbSet[forbidden] {
					isPortForward := len(resources) == 1 && resources[0] == "pods/portforward" && forbidden == "create"
					if !isPortForward {
						t.Errorf("%s Role grants %q on %v", path, forbidden, resources)
					}
				}
			}
		}
		subjects, _ := binding["subjects"].([]any)
		if len(subjects) != 1 {
			t.Fatalf("%s RoleBinding must bind exactly one subject", path)
		}
		subject, _ := subjects[0].(map[string]any)
		if subject["kind"] != "Group" {
			t.Errorf("%s binds kind %v, want Group", path, subject["kind"])
		}
		group, _ := subject["name"].(string)
		if group == "" || !strings.Contains(group, strings.TrimSuffix(owner, "-team")) {
			t.Errorf("%s binds group %q, want the %s developers group", path, group, owner)
		}
		if other, exists := groups[group]; exists {
			t.Errorf("group %q bound in %s and %s", group, other, path)
		}
		groups[group] = path
		roleRef, _ := binding["roleRef"].(map[string]any)
		if roleRef["kind"] != "Role" || roleRef["name"] != "team-developer" {
			t.Errorf("%s roleRef = %v, want the namespaced team-developer Role", path, roleRef)
		}
	}
}
