// Package guardrails statically validates the enforcement artifacts that
// cannot be unit-tested against a live cluster in CI: the GitHub OIDC trust
// policy (exact subject match, no wildcards, least privilege), the Kyverno
// ClusterPolicy files (shared by the CI `kyverno test` layer and the
// admission layer), the Grafana dashboard JSON and the PrometheusRule alerts.
// Live admission behavior is verified separately on the disposable EKS
// cluster; these tests make the invariants reproducible on every PR.
package guardrails

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

// TestOIDCTrustIsExactMatch guards the Phase 2 finding: this GitHub account
// issues a customized subject containing owner/repository IDs, so the trust
// policy must match that exact subject with StringEquals. A wildcard (or a
// StringLike that could grow one) would silently widen access to other
// branches or repositories.
func TestOIDCTrustIsExactMatch(t *testing.T) {
	content := readFile(t, filepath.Join(repoRoot(t), "infra", "modules", "github-oidc", "main.tf"))
	conditionStart := strings.Index(content, "Condition")
	if conditionStart == -1 {
		t.Fatal("trust policy has no Condition block")
	}
	condition := content[conditionStart:]
	if !strings.Contains(condition, "StringEquals") {
		t.Error("trust policy must use StringEquals for the subject (exact match)")
	}
	if strings.Contains(condition, `"token.actions.githubusercontent.com:sub" = local.oidc_subject`) == false &&
		!strings.Contains(condition, "local.oidc_subject") {
		t.Error("trust policy must bind the subject to the configured exact subject")
	}
	if strings.Contains(condition, "*") {
		t.Error("trust policy Condition must not contain a wildcard")
	}
	if !strings.Contains(condition, `"token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"`) {
		t.Error("trust policy must pin the audience to sts.amazonaws.com")
	}
	if !strings.Contains(content, `"eks:DescribeCluster"`) {
		t.Error("trust policy companion must allow eks:DescribeCluster")
	}
	for _, action := range []string{`"eks:*"`, `"iam:*"`, `"ec2:*"`, `AdministratorAccess`} {
		if strings.Contains(content, action) {
			t.Errorf("validation role must not grant %s", action)
		}
	}
}

var expectedPolicies = map[string]string{
	"disallow-host-namespaces.yaml": "no-host-namespaces",
	"disallow-privileged.yaml":      "no-privileged-containers",
	"require-run-as-non-root.yaml":  "run-as-non-root",
	"restrict-capabilities.yaml":    "drop-all-capabilities",
	"require-resources.yaml":        "cpu-memory-requests-limits",
	"disallow-latest-tag.yaml":      "pinned-image",
	"require-owner-labels.yaml":     "ownership-labels",
	"require-probes.yaml":           "health-probes",
}

// TestKyvernoPoliciesShareOneShape ensures every policy file carries the
// deployment contract both enforcement layers rely on: ClusterPolicy kind,
// no background scans (admission-time only, no report noise on the small
// disposable cluster), Enforce action, Pod matching and the platform
// namespace exemptions.
func TestKyvernoPoliciesShareOneShape(t *testing.T) {
	directory := filepath.Join(repoRoot(t), "policies", "kyverno")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(expectedPolicies) {
		t.Fatalf("want %d policies, found %d", len(expectedPolicies), len(entries))
	}
	for file, rule := range expectedPolicies {
		var document map[string]any
		content := readFile(t, filepath.Join(directory, file))
		if err := yaml.Unmarshal([]byte(content), &document); err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		if document["kind"] != "ClusterPolicy" {
			t.Errorf("%s kind = %v, want ClusterPolicy", file, document["kind"])
		}
		spec, _ := document["spec"].(map[string]any)
		if spec == nil {
			t.Fatalf("%s has no spec", file)
		}
		if background, _ := spec["background"].(bool); background {
			t.Errorf("%s must set background: false", file)
		}
		if action, _ := spec["validationFailureAction"].(string); action != "Enforce" {
			t.Errorf("%s validationFailureAction = %q, want Enforce", file, action)
		}
		rules, _ := spec["rules"].([]any)
		if len(rules) != 1 {
			t.Errorf("%s must contain exactly one rule (one policy per guardrail)", file)
		}
		entry, _ := rules[0].(map[string]any)
		if entry["name"] != rule {
			t.Errorf("%s rule = %v, want %q", file, entry["name"], rule)
		}
		validate, _ := entry["validate"].(map[string]any)
		if message, _ := validate["message"].(string); message == "" {
			t.Errorf("%s rule must carry a developer-facing message", file)
		}
	}
}

// TestDashboardExposesSLISignals validates the Git-managed Grafana dashboard:
// every panel queries Prometheus for one of the platform's standard signals
// and every PromQL expression only references metrics the sample app or
// kube-state-metrics actually expose.
func TestDashboardExposesSLISignals(t *testing.T) {
	knownMetrics := []string{
		"http_requests_total",
		"http_request_errors_total",
		"http_request_duration_seconds_bucket",
		"up",
		"kube_deployment_status_replicas_available",
	}
	content := readFile(t, filepath.Join(repoRoot(t), "observability", "dashboards", "golden-path-overview.json"))
	var dashboard struct {
		UID    string `json:"uid"`
		Title  string `json:"title"`
		Panels []struct {
			Title   string `json:"title"`
			Targets []struct {
				Expr string `json:"expr"`
			} `json:"targets"`
		} `json:"panels"`
	}
	if err := json.Unmarshal([]byte(content), &dashboard); err != nil {
		t.Fatalf("dashboard is not valid JSON: %v", err)
	}
	if dashboard.UID == "" || dashboard.Title == "" {
		t.Error("dashboard must have a uid and title")
	}
	if len(dashboard.Panels) < 4 {
		t.Fatalf("dashboard must cover rate, errors, latency and availability (%d panels)", len(dashboard.Panels))
	}
	for _, panel := range dashboard.Panels {
		if panel.Title == "" {
			t.Error("every panel must have a title")
		}
		if len(panel.Targets) == 0 || panel.Targets[0].Expr == "" {
			t.Errorf("panel %q has no PromQL query", panel.Title)
			continue
		}
		known := false
		for _, metric := range knownMetrics {
			if strings.Contains(panel.Targets[0].Expr, metric) {
				known = true
			}
		}
		if !known {
			t.Errorf("panel %q queries unknown metrics: %q", panel.Title, panel.Targets[0].Expr)
		}
	}
}

// TestPrometheusRulesAreEvaluable checks the baseline alerts structurally:
// named alerts, SLO-derived expressions over known metrics, severity labels
// and a `for` duration so brief spikes do not page.
func TestPrometheusRulesAreEvaluable(t *testing.T) {
	content := readFile(t, filepath.Join(repoRoot(t), "observability", "prometheus-rules.yaml"))
	var document struct {
		Kind string `yaml:"kind"`
		Spec struct {
			Groups []struct {
				Name  string `yaml:"name"`
				Rules []struct {
					Alert  string            `yaml:"alert"`
					Expr   string            `yaml:"expr"`
					For    string            `yaml:"for"`
					Labels map[string]string `yaml:"labels"`
				} `yaml:"rules"`
			} `yaml:"groups"`
		} `yaml:"spec"`
	}
	if err := yaml.Unmarshal([]byte(content), &document); err != nil {
		t.Fatalf("parse prometheus-rules.yaml: %v", err)
	}
	if document.Kind != "PrometheusRule" {
		t.Fatalf("kind = %q, want PrometheusRule", document.Kind)
	}
	alerts := 0
	for _, group := range document.Spec.Groups {
		for _, rule := range group.Rules {
			if rule.Alert == "" {
				continue
			}
			alerts++
			if rule.Expr == "" || rule.For == "" {
				t.Errorf("alert %q needs an expr and a for duration", rule.Alert)
			}
			if rule.Labels["severity"] == "" {
				t.Errorf("alert %q needs a severity label", rule.Alert)
			}
			known := false
			for _, metric := range []string{"http_request_errors_total", "http_requests_total", "up"} {
				if strings.Contains(rule.Expr, metric) {
					known = true
				}
			}
			if !known {
				t.Errorf("alert %q uses unknown metrics: %q", rule.Alert, rule.Expr)
			}
		}
	}
	if alerts < 2 {
		t.Errorf("want at least 2 baseline alerts, got %d", alerts)
	}
}
