# Validation record

Status meanings:

- **NOT RUN**: no execution evidence exists.
- **PASS**: observed result matched the stated expectation.
- **PARTIAL**: useful evidence exists but part of the claim was not verified.
- **FAIL**: observed result contradicted the expectation.

This file is updated only from observed command output. AWS and hosted GitHub
Actions remain **NOT RUN** in Phase 2.

## Phase 2 local validation

Executed again after Docker recovery on 2026-09-22 JST on Windows. Helm, kubeconform and Trivy were run
from temporary, version-pinned downloads; they were not installed globally.

Versions: Go 1.27.0, Terraform 1.14.3, AWS provider 6.66.0, TLS provider 4.4.1,
Helm 3.18.6, kubectl 1.36.1, kubeconform 0.6.7, and Trivy 0.74.0.

| Check | Command | Observed result | Status |
| --- | --- | --- | --- |
| Go CLI tests | `go test ./...` | command and contract packages passed | PASS |
| Go CLI static analysis | `go vet ./...` | no findings | PASS |
| Sample tests/static analysis | `cd sample-app && go test ./... && go vet ./...` | health, readiness and metrics handler test passed; vet had no findings | PASS |
| CLI build and contract | `go build -o platform.exe ./cmd/platform` then `platform.exe validate services/.../service.yaml` | `PASS services\\payments-team\\payment-api\\service.yaml` | PASS |
| Doctor | `platform.exe doctor` with temporary Helm directory on `PATH` | git, go, terraform, helm and kubectl found | PASS |
| Terraform formatting | `terraform -chdir=infra fmt -check -recursive` | exit 0 | PASS |
| Terraform initialization/validation | `terraform -chdir=infra init -backend=false` and `terraform -chdir=infra validate` | providers installed/reused; `Success! The configuration is valid.` | PASS |
| Helm lint/render | `helm lint ...` and two `helm template` runs | standard path rendered 3 resources; HPA path rendered 4 | PASS |
| Kubernetes schema | `kubeconform -strict -summary -kubernetes-version 1.33.0 rendered.yaml rendered-hpa.yaml` | 7 valid, 0 invalid, 0 errors, 0 skipped | PASS |
| GitOps rendering | `kubectl kustomize gitops/platform` | exit 0 | PASS |
| Shell syntax | `bash -n scripts/bootstrap-gitops.sh scripts/destroy.sh scripts/local-validate.sh` | exit 0 | PASS |
| Filesystem security scan | `trivy fs --scanners vuln,secret,misconfig --severity HIGH,CRITICAL --exit-code 1 --skip-dirs infra/.terraform .` | 0 dependency vulnerabilities, secrets, and remaining HIGH/CRITICAL misconfigurations | PASS |
| Linux sample binary | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ...` | 5,828,768-byte linux/amd64 binary created in the temporary directory | PASS |
| Sample container image | `docker build -t platform-sample-app:local sample-app` | final rerun built image `sha256:64d2f5f7189aae73aa3b37cf49a00f02bb6386631b241235f03f0e665a0a2990`; container `/healthz` returned HTTP 200 with body `ok` | PASS |
| Hosted GitHub Actions | push/PR workflow | repository is not published from this workspace | NOT RUN |
| AWS/EKS apply | approval-gated real validation | no AWS API mutation performed | NOT RUN |
| Argo CD reconciliation | real cluster validation | no cluster exists | NOT RUN |

### Security scan disposition

The first Trivy run found six infrastructure findings. Two public-subnet auto-IP
findings were fixed, and unrestricted EKS API CIDRs were replaced with a required
validated input that rejects `0.0.0.0/0`. Three inline exceptions remain and are
visible in code:

- public EKS API access is needed for the remote disposable validation step and
  is restricted to caller-supplied CIDRs plus IAM authentication;
- private node egress is needed through the NAT for image and Git access;
- EKS 1.28+ applies default envelope encryption with an AWS-owned KMS key, so a
  paid customer-managed key was not added without a compliance requirement.

Trivy was rerun after these changes and returned exit 0. Scanner exceptions are
design decisions, not proof that AWS configuration behaves correctly.

### Local verdict

**PASS** for the defined local validation scope. Docker Desktop was repaired and
the image build plus a real container health request passed. No hosted CI,
cluster, GitOps reconciliation, or AWS behavior is promoted to PASS.
