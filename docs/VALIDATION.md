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
HA/failure injection (addressed in Phase 3 below where stated).

## Phase 3 validation (2026-09-22 JST, second disposable environment)

Phase 3 implementation (commit `6ad432f`) was validated locally in full, then
against a fresh disposable environment in `ap-northeast-1`
(`developer-platform-dev`, EKS 1.35, `t3.medium` x2, 1 NAT, no RDS/LB/mesh/GPU).
Account IDs are masked as `<masked-account>`. All timestamps JST unless noted.

### P3-1. Local validation (all PASS, 2026-09-22 ~21:00-21:25 JST)

Go 1.27.0, Terraform 1.14.3, Helm 3.18.6, kubeconform 0.6.7, Kyverno CLI
v1.19.1, Trivy 0.74.0 (pinned temp downloads), kubectl 1.36.1, Docker.

| Check | Command | Observed | Status |
| --- | --- | --- | --- |
| Go tests (all packages) | `go test ./...` | command, contract, gitops, goldenpath, guardrails ok | PASS |
| Go vet | `go vet ./...` + sample-app | no findings | PASS |
| Contract both teams | `platform validate services/...` x2 | PASS x2 (incl. environment/contact, digest pins) | PASS |
| Terraform | fmt/init/validate | exit 0 / Success | PASS |
| Helm lint/render both teams | lint x2, template x3 | 0 failed; payments 3, orders 3, HPA 4 resources | PASS |
| kubeconform 1.35.0 | 3 rendered files | 10 valid, 0 invalid/errors/skipped | PASS |
| Kyverno suite | `kyverno test policies/tests/` | 40 passed, 0 failed (cases A-E) | PASS |
| Kustomize | `kubectl kustomize gitops/platform` | exit 0 | PASS |
| Dashboard JSON | python json.load | valid | PASS |
| Shell syntax | `bash -n` x5 scripts | exit 0 | PASS |
| Trivy | fs HIGH,CRITICAL exit 1 | exit 0 (see disposition) | PASS |
| Docker + /metrics | build, run, curl | HTTP 200, counter+histogram exposition verified | PASS |

Trivy disposition: one HIGH (`KSV-0017`) fired on
`policies/tests/invalid-privileged.yaml` — the intentional negative fixture
(asserted REJECT by `kyverno test`). True positive on test data, not a
product finding: the fixture directory is skipped (`--skip-dirs
infra/.terraform,policies/tests`), documented in `local-validate.sh` and CI.
A Helm-scanner warning also exposed that `charts/golden-path/values.yaml`
(defaults) missed the new required `environment` key; fixed by adding it.

### P3-2. Terraform plan review and apply

- Status: **PASS**
- Plan (2026-09-22 ~21:30 JST): `42 add / 0 change / 0 destroy` —
  Phase 2 shape plus the carried-over node-to-API 443 SG rule. Verified: region
  `ap-northeast-1`, EKS 1.35, `t3.medium` desired 2/max 3, 1 NAT + 1 EIP, 1
  ECR (IMMUTABLE), CW log group 7d retention, project-owned OIDC
  provider/role/policy, SG rules incl. the 443 fix. No RDS/LB/GPU/mesh.
- Apply: `Apply complete! Resources: 42 added, 0 changed, 0 destroyed.`
  No plan/apply drift. Two follow-up applies during validation added 2 rules
  (`nodes_from_eks_managed_sg` 443/10250) then 1 port (`9443`); see P3-4.

### P3-3. EKS and container image

- Status: **PASS**
- `eks list-clusters` shows `developer-platform-dev`; nodes 2/2 Ready
  (`v1.35.8-eks-a887778`); add-ons `vpc-cni/coredns/kube-proxy` ACTIVE;
  system Pods Running, 0 restarts (21:55 JST snapshot).
- Image: `docker build` + push of two immutable tags (`payment-api-p3`,
  `order-api-p3`), both digest
  `sha256:b359742b13d36e0574dcfe492f38c6367c2e90c9ed7d0e4c23f6acc6c2fcfe05`;
  both Service Definitions pinned to the digest (commit `d2e0b2a`), contract
  validation PASS on both.

### P3-4. Failure: admission webhook unreachable (new incident)

- Symptom: `kubectl apply -f policies/kyverno/` failed on all 8 files:
  `failed calling webhook "mutate-policy.kyverno.svc": ... context deadline exceeded`.
- Investigation: kyverno-svc endpoints existed (`10.40.11.125:9443`),
  controller Running; node SG allowed 443/10250 only from the custom cluster
  SG, while control-plane ENIs carry the EKS-managed SG. First fix added
  443/10250 from the managed SG (`vpc_config[0].cluster_security_group_id`) —
  still failing. Second look: the webhook Pod serves **9443** (Service 443
  maps to targetPort 9443), and pod-port traffic is filtered by the node SG.
- Root cause: two gaps — (1) managed control-plane SG not admitted to nodes,
  (2) webhook pod port 9443 not opened. Phase 2 never hit this because Argo CD
  installs no admission webhooks.
- Fix (commit `0dc91ba`): `nodes_from_eks_managed_sg` allowing
  443/10250/9443 SG-to-SG from the EKS-managed cluster SG. No CIDR widening,
  no `0.0.0.0/0`. Re-applied (2 + 1 adds), policies created: 8/8 Ready.
- Re-validation: full admission battery P3-5. Security was not weakened to
  route around the problem.
- Note: the bootstrap script's version assertion initially read the wrong
  label (`app.kubernetes.io/version` = chart version); corrected to assert
  the deployment image tag (`:v1.19.1`).

### P3-5. Policy admission (live)

- Status: **PASS** (2026-09-22 ~21:45 JST)
- Invalid Pods A-D via `kubectl apply --dry-run=server`: all REJECTED with
  `admission webhook "validate.kyverno.svc-fail" denied the request`.
- Valid Pod fixture: `created (server dry run)` (admitted).
- Rendered valid `order-api` Deployment: PDB/Service/Deployment all
  `created (server dry run)` — proves Kyverno autogen covers controllers.
- Rendered Deployment with image rewritten to `nginx:latest`: PDB/Service
  pass, Deployment REJECTED — autogen denial at the controller level.
- Pre-existing workloads (deployed before policies) untouched; no background
  scan noise (`background: false`).

### P3-6. Multi-team and GitOps

- Status: **PASS**
- Argo CD 8.3.0 bootstrapped; `platform-root`, `payment-api`, `order-api`
  all `Synced/Healthy` without manual intervention (namespace derivation
  `team-{{owner}}` worked for both teams).
- `team-payments/payment-api` 3/3 Ready, `team-orders/order-api` 2/2 Ready,
  0 restarts, ClusterIP Services + PDBs present, live images digest-pinned
  (`@sha256:b359...`).
- `/healthz` via port-forward at 21:49 JST: both HTTP 200 body `ok`.

### P3-7. RBAC boundary (live, impersonated SelfSubjectAccessReview)

- Status: **PASS**
- payments group in `team-payments`: `get deployments` yes, `get pods/log` yes.
- payments group in `team-orders`: get deployments **no**, get pods **no**.
- payments group `delete deployments` in own namespace: **no** (GitOps owns writes).
- orders group: yes in `team-orders`, **no** in `team-payments`.
- Subjects are IdP group names; no user credentials in Git.

### P3-8. Observability (live, real traffic)

- Status: **PASS** (2026-09-22 ~21:51-21:54 JST)
- Stack: kube-prometheus-stack 91.4.1 installed with minimal values (no
  Alertmanager); all pods Running; `golden-path-slo` PrometheusRule created.
- Traffic: 120 requests to each service root (+ kubelet probes).
- Prometheus (`up{job="golden-path"}`): 5/5 targets up with `service` label
  (annotation scrape + relabeling works).
- SLI queries against real data: request rate approx 0.41/s; availability
  `= 1`; p95 approx 4.75ms (SLO: 99% within 500 ms — headroom confirmed live).
- Grafana: `golden-path-overview` dashboard provisioned from Git (API search
  + fetch: 5 panels, queries identical to the verified SLI queries).
  Limitation: panel data verified via Prometheus API, not via in-browser
  screenshot (no browser in the validation environment); the Grafana
  `/api/ds/query` time-range plumbing was not cracked, recorded as a minor
  tooling gap, not a platform gap.

### P3-9. GitHub OIDC positive and negative

- Status: **PASS**
- Positive (main, run `35730184574`, all 7 jobs green incl. new `policy`
  job): `sts get-caller-identity` assumed the validation role and
  `describe-cluster` returned the live cluster — with the trust hardened to
  `StringEquals` on the exact custom subject.
- Negative (branch `tmp/oidc-negative`, run `35729948912`, workflow_dispatch):
  subject `...:ref:refs/heads/tmp/oidc-negative` → `configure-aws-credentials`
  retried and FAILED with access denied; no credentials issued. Branch deleted
  locally and remotely afterwards. No wildcard was added; the design is
  unchanged.
- Post-destroy hygiene: `AWS_VALIDATION_ROLE_ARN` variable deleted again so
  later runs skip the job instead of failing against the removed role.

### P3-10. CI finding: cluster-dependent check in manifests job

- Symptom: `kubectl apply --dry-run=client -f prometheus-rules.yaml` failed
  in CI (`connection refused` — even client dry-run needs API discovery for
  CRDs).
- Fix (commit `ca1343e`): removed the kubectl line; rule content is asserted
  in Go (`internal/guardrails`, runs in the `go` job). Re-run green.

### P3-11. Destroy and residuals (2026-09-22 22:05-22:10 JST)

- Status: **PASS**
- Destroy plan: `0 add / 0 change / 45 destroy` (42 + 2 managed-SG rules +
  1 port) — all project-owned. `Apply complete! ... 45 destroyed.`
- `infra/terraform.tfstate`: 0 resources. `eks list-clusters`: empty.
  Project VPC/subnets/NAT/EIP/ENI/SG/EBS/ECR/CW-logs/IAM-roles/OIDC-provider:
  all absent (EC2 tombstones converged to none). No LoadBalancer was ever
  created. No pre-existing AWS resource touched.

### P3-12. Cost

No billing-console measurement (limitation). About 1h-lived environment:
EKS control plane, 2x `t3.medium`, 1 NAT + EIP, EBS, ECR, CloudWatch —
plus the small monitoring/Kyverno/Argo CD footprints on the same nodes (no
extra instances). No RDS/LB/GPU/mesh. No exact figure claimed.

### Phase 3 verdict

| Area | Status |
| --- | --- |
| Implementation (guardrails/multi-team/RBAC/observability/SLO/OIDC-neg) | PASS |
| Local validation (existing + new) | PASS |
| Real AWS infrastructure | PASS |
| Policy admission | PASS |
| Multi-team / GitOps / app | PASS |
| RBAC boundary | PASS |
| Observability / dashboard / SLI | PASS |
| GitHub OIDC positive + negative | PASS |
| Destroy | PASS |
| Residual resource check | PASS |

Remaining limitations: 30-day SLO compliance cannot be proven by a 1-hour
cluster (queries evaluate, budget burn does not); Grafana panel rendering
verified via API + Prometheus data, not screenshot; init/ephemeral containers
outside policy scope; ClusterPolicy deprecation noted with ValidatingPolicy
migration as follow-up; no log aggregation (by design); namespace is not a hard
security boundary (documented).

## Phase 4 reliability validation (2026-09-23 JST)

Fresh disposable environment: Terraform `45 add / 0 change / 0 destroy`, EKS
1.35 ACTIVE, two nodes Ready, Argo/Kyverno/Prometheus bootstrapped, and baseline
root/payments/orders Synced+Healthy at commit `46764c0`. Both health endpoints
returned 200; 5/5 Golden Path targets were UP; availability query returned 1.

| Area | Result |
| --- | --- |
| A — policy-valid bad image | PARTIAL: old 3 replicas preserved 50/50 HTTP 200; Argo refresh required manual hard refresh; Git fix recovered Healthy |
| B — Argo controller loss | PARTIAL: application remained available; deploy/drift/status stopped; recovery required hard refresh; managed replica drift self-healed |
| C — Prometheus/operator loss | PASS: app 50/50 HTTP 200 while queries/SLI unavailable; 5/5 targets and SLI returned after recovery |
| D — team RBAC boundary | PASS: payments list/log/port-forward worked; orders deployment server-dry-run create was Forbidden |
| Optional node maintenance | NOT RUN: intentionally omitted to avoid duplicating Kubernetes rescheduling mechanics |

Detailed pre-written hypotheses, timestamps, SHAs, observations, recovery, and
differences are in `docs/FAILURE_EXPERIMENTS.md`. The short samples are not
production availability or recovery benchmarks.

### Phase 4 destroy and final local validation

- Destroy: **PASS**, plan `0/0/45`, apply `45 destroyed`, state count 0.
- Residual check: **PARTIAL**. Every direct API category was empty (EKS,
  EC2/ASG, VPC/subnet/NAT/EIP/ENI/SG/EBS, ECR, CloudWatch, IAM roles,
  project-created OIDC, load balancers), but Resource Groups Tagging API still
  returned the deleted NAT ARN after retries.
- Final local validation: **PASS**. Go tests/vet, contracts, Terraform,
  Helm, kubeconform (10/10), GitOps render, observability, Kyverno cases A–E,
  shell syntax, Trivy source scan, Docker build, and container `/healthz` 200.
- Cost: no billing total claimed; the EKS/two-node/NAT environment existed for
  approximately 80 minutes.
