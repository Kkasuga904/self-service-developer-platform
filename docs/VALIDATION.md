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

## Real AWS / EKS validation (2026-09-22 JST, disposable environment)

Two sessions share this record. A first session performed plan, apply, EKS /
ECR / Argo CD / GitOps / drift / OIDC validation, then started destroy. A
resume session (this one) directly observed destroy completion, ran the full
residual-resource verification, re-ran every local validation, and fetched the
hosted CI logs. Live-cluster checks that could no longer be re-observed after
deletion are marked **PARTIAL** with the limitation stated; nothing is promoted
to PASS without observed evidence.

Conventions below: account IDs are masked as `<masked-account>`. No
credentials, tokens, kubeconfig material, OIDC JWTs, or full Terraform state
are stored in this repository.

Environment: region `ap-northeast-1`, name prefix `developer-platform-dev`,
EKS `1.35`, nodes `t3.medium` (desired 2, max 3), 1 NAT Gateway, 1 EIP, 1 ECR
repository (`IMMUTABLE`, scan-on-push, keep-10 lifecycle), 1 CloudWatch log
group, project-owned GitHub OIDC provider (created because
`existing_github_oidc_provider_arn` was empty), no RDS, no LoadBalancer, no
service mesh, no GPU. Head commit for this record:
`5008a71e5e029f53d7f42a9b4dad448018b9efe5`.

### 1. EKS version pin (observed at resume)

- Status: **PASS**
- Command: `aws eks describe-cluster-versions --region ap-northeast-1`
  (profile `portfolio`, 2026-09-22 ~17:38 JST)
- Expected: a standard-support version, not an extended-support one.
- Observed: `STANDARD_SUPPORT` = `1.34`, `1.35`, `1.36`; default = `1.36`.
- Decision recorded:   pinned `1.35` (`infra/variables.tf` default,
  `infra/terraform.tfvars`, CI kubeconform `-kubernetes-version 1.35.0`).
  Reason: `1.35` keeps a long standard-support window while avoiding day-zero
  churn of the newest default (`1.36`). Kubernetes `1.33` (extended support)
  is not used anywhere.
- Limitation: standard-support end date was reported by the prior session as
  2027-03-27 JST; the resume session confirmed the support *status*, not the
  calendar date.

### 2. Local validation re-run at resume (all PASS)

Executed 2026-09-22 ~17:38-17:45 JST on Windows. Helm 3.18.6, kubeconform
0.6.7 and Trivy 0.74.0 from the existing version-pinned temporary downloads
(not installed globally). Go 1.27.0, Terraform 1.14.3, kubectl 1.36.1.

| Check | Command | Observed result | Status |
| --- | --- | --- | --- |
| Go CLI tests | `go test ./...` | command and contract packages passed | PASS |
| Go CLI static analysis | `go vet ./...` | no findings | PASS |
| Sample tests/static analysis | `cd sample-app && go test ./... && go vet ./...` | passed, no findings | PASS |
| CLI build and contract | `go build -o platform.exe ./cmd/platform` then `platform.exe validate services/payments-team/payment-api/service.yaml` | `PASS services\payments-team\payment-api\service.yaml` (digest-pinned image, replicas 3) | PASS |
| Terraform formatting | `terraform -chdir=infra fmt -check -recursive` | exit 0 | PASS |
| Terraform init/validate | `terraform -chdir=infra init -backend=false`, `terraform -chdir=infra validate` | `Success! The configuration is valid.` | PASS |
| Helm lint/render | `helm lint charts/golden-path -f services/.../service.yaml`, two `helm template` runs | 0 failed; standard 3 resources, HPA 4 | PASS |
| Kubernetes schema | `kubeconform -strict -summary -kubernetes-version 1.35.0 rendered.yaml rendered-hpa.yaml` | 7 valid, 0 invalid, 0 errors, 0 skipped | PASS |
| GitOps rendering | `kubectl kustomize gitops/platform` | exit 0 | PASS |
| Shell syntax | `bash -n scripts/bootstrap-gitops.sh scripts/destroy.sh scripts/local-validate.sh` (WSL) | exit 0 | PASS |
| Filesystem security scan | `trivy fs --scanners vuln,secret,misconfig --severity HIGH,CRITICAL --exit-code 1 --skip-dirs infra/.terraform .` | exit 0 after stale-artifact disposition below | PASS |
| Sample container image + health | `docker build -t platform-sample-app:local sample-app`, `docker run`, `curl /healthz` | HTTP 200, body `ok` (2026-09-22 17:41 JST) | PASS |

Trivy disposition at resume: the first rerun failed with one CRITICAL
(`AWS-0041`, public cluster access) raised against the *plan snapshot* of the
git-ignored stale local file `infra/tfplan` (post-destroy leftover), flagging
the operator's single trusted `/32` CIDR. The Terraform *sources* carry an
explicit inline exception for this design decision (public endpoint restricted
to caller-supplied CIDRs plus IAM auth, never `0.0.0.0/0`, which the variable
validation rejects). Fix: deleted the ignored stale artifacts (`infra/tfplan`,
regenerated `rendered*.yaml`); rerun exit 0. No security posture was weakened.
The three pre-existing inline exceptions (restricted public API access,
private-node NAT egress, AWS-owned default envelope encryption instead of a
paid customer-managed KMS key) are unchanged.

### 3. Terraform plan review and apply

- Status: **PARTIAL** (plan/apply ran in the first session; the resume session
  verifies the outcome, not the live transcript).
- Reported plan (first session): `41 add / 0 change / 0 destroy`,
  `ap-northeast-1`, EKS `1.35`, `t3.medium` x2 (max 3), 1 NAT, 1 EIP, 1 ECR,
  1 CloudWatch log group, 1 project-owned OIDC provider; no RDS, LoadBalancer,
  mesh, multi-AZ NAT, or GPU. All resources Terraform-managed and destroyable.
- Reported apply: VPC, subnets, NAT, IAM, EKS, managed node group, 3 add-ons,
  ECR, GitHub OIDC resources created from the saved plan with no material
  plan/apply drift.
- Corroboration observable at resume: `infra/terraform.tfstate` lists 0
  resources after destroy; `terraform.tfstate.backup` (ignored, local only)
  retains the pre-destroy snapshot; service definition pins the ECR digest
  produced by the apply (`.../developer-platform-dev/payment-api@sha256:953e...fcae`,
  full digest in `services/payments-team/payment-api/service.yaml`); CI run
  `35704359374` logged `describe-cluster` = `ACTIVE / 1.35`.
- Limitation: the live plan transcript and apply log are not re-observable
  post-destroy.

### 4. EKS validation

- Status: **PARTIAL**.
- Corroborating evidence: CI OIDC job (run `35704359374`,
  `2026-09-22T08:21:44Z`) observed `developer-platform-dev | ACTIVE | 1.35`
  via `aws eks describe-cluster`. Prior session reported nodes Ready, 3
  add-ons (`vpc-cni`, `coredns`, `kube-proxy`) ACTIVE, system Pods healthy
  after the SG fix in section 7.
- Limitation: no live cluster exists at resume (destroyed); `kubectl` checks
  were not re-observable.

### 5. Container image (ECR, immutable, digest-pinned)

- Status: **PARTIAL**.
- Observed at resume: `services/payments-team/payment-api/service.yaml` pins
  `4407...` ECR image by digest
  (`sha256:953ef4d7f9522a8f98fab183b67cd9270216379cd00a6b23841554b7739cfcae`);
  repository was `IMMUTABLE` with scan-on-push per `infra/main.tf`; CI `image`
  job (docker build, no push) passed in run `35704359374`. The repository
  itself was already destroyed (`RepositoryNotFoundException` on
  `describe-images`, 2026-09-22 17:36 JST), consistent with a complete destroy.
- Limitation: the ECR digest listing and push transcript come from the first
  session and git history, not from a live registry at resume.

### 6. Argo CD / GitOps validation

- Status: **PARTIAL**.
- Observed at resume: manifests render (`kubectl kustomize gitops/platform`
  exit 0); `gitops/platform/` (AppProject, team namespace, ApplicationSet)
  and `gitops/bootstrap/root-application.yaml` are committed (commit
  `5a052bf`); replica change `2 -> 3` is committed (`692bc62`) and the live
  image is digest-pinned, so the raw-Deployment escape hatch was not used.
- Limitation: live Argo CD `Synced/Healthy` states were reported by the first
  session and are not re-observable post-destroy.

### 7. Real application validation (`/healthz`)

- Status: **PARTIAL**.
- Observed at resume: local container built from the same `sample-app`
  answered `/healthz` with HTTP 200 and body `ok` (17:41 JST). The deployed
  manifest path (digest-pinned image, probes on `/healthz`/`/readyz`, PDB,
  ClusterIP Service) is rendered and schema-valid.
- Limitation: the in-cluster Pod/Service HTTP request was reported by the
  first session and is not re-observable post-destroy; "Pod Ready implies
  reachable" is not claimed.

### 8. GitOps change validation (replicas 2 -> 3)

- Status: **PARTIAL**.
- Observed at resume: git history shows the safe value change and its intent
  (`692bc62 Validate GitOps replica change` on top of `5634390`);
  `service.yaml` at HEAD carries `replicas: 3` with the digest-pinned image.
- Limitation: the Argo detection -> sync -> live-replica timeline was reported
  by the first session and is not re-observable post-destroy.

### 9. Drift / self-heal experiment

- Status: **PARTIAL**.
- Basis: first-session record (manual `scale --replicas=1` against a Git
  desired state of 3, `OutOfSync` observed after forcing a hard refresh,
  self-heal back to 3/3 `Synced/Healthy`).
- Limitation: the second-by-second timeline is not re-observable post-destroy;
  only the methodology and outcome are recorded here, not re-verified.

### 10. GitHub Actions OIDC (directly observed at resume)

- Status: **PASS**.
- Run: `35704359374` (`Support exact customized GitHub OIDC subjects`,
  push on `main`, head `5008a71`), conclusion `success`, all 6 jobs green
  (`go`, `terraform`, `manifests`, `security`, `image`, `aws-oidc`).
  Observed via `gh run view` and the `aws-oidc` job log (ID `106669603389`)
  on 2026-09-22 ~17:37 JST.
- Expected vs observed:
  - OIDC federation: PASS. Token claims logged without secrets:
    `sub=repo:Kkasuga904@<owner-id>/self-service-developer-platform@<repo-id>:ref:refs/heads/main`,
    `aud=sts.amazonaws.com`, `repository=Kkasuga904/self-service-developer-platform`,
    `ref=refs/heads/main`.
  - Expected IAM role: PASS. Assumed
    `arn:aws:sts::<masked-account>:assumed-role/developer-platform-dev-github-validation/GitHubActions`
    via `aws sts get-caller-identity`.
  - Least privilege: PASS. Only `sts get-caller-identity` and
    `eks describe-cluster --name developer-platform-dev` ran from the
    workflow; the latter returned `ACTIVE / 1.35`.
  - No long-lived credentials: PASS. `gh secret list` is empty; auth used
    `aws-actions/configure-aws-credentials@v5` with `id-token: write`.
  - Repo scoping: the trust policy binds the exact custom subject above
    (`infra/terraform.tfvars` + `github_oidc_subject` support, commit
    `5008a71`).
- Post-destroy hygiene: the role no longer exists, so the non-secret repo
  variable `AWS_VALIDATION_ROLE_ARN` (which pointed at it) was deleted at
  resume; future runs skip the `aws-oidc` job instead of failing against a
  deleted role.
- Limitation: the intentional negative test (unauthorized branch denied) was
  not performed; sent to Phase 3.

### 11. Failure handling (first-session record, preserved verbatim in substance)

1. Private nodes never joined (`EC2 running` but node group `CREATING`
   >12 min). Symptom: node group stuck. Root cause: the extra cluster
   security group lacked TCP 443 ingress from the node security group, so
   nodes using the private endpoint could not reach the API. Fix (commit
   `4166a53`): minimal Terraform SG rule allowing node SG ->
   control-plane SG on 443 only. Re-validation: re-apply, nodes Ready,
   add-ons ACTIVE. Security was not weakened (no `0.0.0.0/0`).
2. First apply wait cancelled safely (no forced AWS deletion); unregistered
   node group imported to restore Terraform ownership, then re-applied.
   Recorded as FAIL-then-fixed, not hidden.
3. Argo CD root Application `ComparisonError: gitops/platform: app path does
   not exist`. Root cause: unanchored `platform` entry in `.gitignore`
   excluded `cmd/platform/` and `gitops/platform/`. Fix (commits `5a052bf`,
   `5634390`): anchored ignore rules, committed the missing manifests.
4. CI `security` job failed: `aquasecurity/trivy-action@0.33.1` does not
   exist. Fix (commit `2c9de72`): released tag `v0.36.0` with pinned
   `trivy-version v0.74.0`.
5. OIDC `AssumeRoleWithWebIdentity` denied: this GitHub account emits a
   custom subject containing owner/repository IDs, not the standard
   `repo:OWNER/NAME:ref:...` form. Lowercase normalization was tried and was
   wrong. Fix (commit `5008a71`): `github_oidc_subject` variable carrying the
   observed exact subject, strictly repository/main-scoped, no wildcards.

### 12. Destroy (directly observed at resume)

- Status: **PASS**.
- Observed 2026-09-22 17:36 JST: no `terraform` process remains;
  `infra/terraform.tfstate` parses to **0 resources**; `eks list-clusters`
  returns `[]`; the earlier eventually-consistent `describe-cluster` /
  `describe-nodegroup` `ResourceNotFoundException` converged to `list-clusters
  == []`.
- Reported destroy plan (first session): `0 add / 0 change / 42 destroy`,
  covering only project-owned resources including the project-created GitHub
  OIDC provider. No existing VPC, IAM role, EKS, OIDC provider, ECR,
  CloudWatch, Route53, Organizations, IAM user, or account-level security
  configuration was touched (existing-provider ARN input stayed empty, so the
  provider Terraform deleted was the one it created).

### 13. Residual resource verification (directly observed at resume, 2026-09-22 17:36-17:37 JST, `ap-northeast-1`, profile `portfolio`)

| Resource | Command | Observed | Status |
| --- | --- | --- | --- |
| EKS | `eks list-clusters` | `[]` | PASS |
| Managed Node Group / ASG | `describe-nodegroup` (`ResourceNotFoundException`), ASG name filter | none | PASS |
| EC2 | `describe-instances` filter `eks:cluster-name=developer-platform-dev` | 2 `terminated` tombstones, none running | PASS |
| VPC / Subnets | tag `Project=self-service-developer-platform` | none | PASS |
| NAT Gateway / EIP | tag filter + address query | none active (one `deleted` tombstone seen mid-destroy) | PASS |
| ENI / Security Groups | tag filter | none | PASS |
| EBS | tag filter | none | PASS |
| ECR | `describe-repositories` filter `developer-platform-dev` | `[]` (`RepositoryNotFoundException` for the sample repo) | PASS |
| CloudWatch Log Group | `describe-log-groups --log-group-name-prefix /aws/eks/developer-platform-dev` | `[]` | PASS |
| IAM Roles | `iam get-role` x3 (cluster, nodes, github-validation) | `NoSuchEntity` x3 | PASS |
| GitHub OIDC Provider | `iam list-open-id-connect-providers` | empty (project-created provider deleted; no pre-existing provider was adopted) | PASS |

State-empty alone was not accepted as completion; every API check above was
executed. Asynchronous `deleting` states converged to absent during the
resume session.

### 14. Cost

No billing-console measurement was taken (limitation). Bounded design for the
~2-hour-lived environment on 2026-09-22 JST: EKS control plane hourly, 2x
`t3.medium` nodes, 1 NAT Gateway + EIP, EBS volumes, ECR storage/requests,
CloudWatch ingestion for the control-plane log group. No RDS, LoadBalancer,
mesh, multi-AZ NAT, or GPU was created. No cost anomaly was observed, but no
exact figure is claimed.

### Real-AWS verdict

| Area | Status |
| --- | --- |
| Local Validation | PASS |
| Real AWS Infrastructure | PARTIAL |
| EKS | PARTIAL |
| Application | PARTIAL |
| GitOps | PARTIAL |
| Drift / Self-heal | PARTIAL |
| GitHub OIDC | PASS |
| Destroy | PASS |
| Residual Resource Check | PASS |

PARTIAL here means: created and exercised in the first session with git/CI
corroboration still observable, but live re-observation is impossible after
the (completed, verified) destroy. The one directly re-observable hosted
proof, GitHub OIDC, is PASS. Deliberately unvalidated: negative-branch OIDC
test, billing figures, admission policy, observability, multi-team isolation,
HA/failure injection (Phase 3+).
