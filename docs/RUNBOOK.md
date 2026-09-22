# Runbook

## Local contract failure

Run `./platform validate <path>`. Correct every reported field; do not bypass
the validator by editing rendered manifests because Argo CD renders from the
contract again.

## Policy rejection in CI

`kyverno test policies/tests/` names the failing case (A–E) and rule. Fix the
workload or, if the Golden Path cannot express a legitimate need, propose a
typed contract field — not a raw override. Do not weaken the policy to make
CI green; weakening a guardrail requires an ADR.

## Argo Application is OutOfSync

This procedure was validated on the disposable cluster in Phase 2:

1. Inspect `argocd app get <service>` and related Kubernetes events.
2. Compare the Application source revision with main.
3. Render the exact values locally with Helm.
4. Fix desired state in Git; avoid an imperative patch.
5. Sync and record the revision and observed recovery.

## Admission rejection on the cluster

1. Read the denial message: it names the ClusterPolicy and rule.
2. Check whether the workload came through the Golden Path; if yes, the
   chart or contract is at fault — fix the platform, not the policy.
3. Hand-written manifests are rejected by design; convert them to a Service
   Definition.
4. Namespace exemptions (`kube-system`, `argocd`, `kyverno`, `monitoring`)
   cover platform components only; never add a team namespace to them.

## Observability quick checks

- Prometheus: port-forward the `kube-prometheus-stack-prometheus` Service in
  `monitoring` and query `up{job="golden-path"}`.
- SLI spot check: evaluate the availability and p95 queries from docs/SLO.md
  against the last 5m.
- Grafana: admin password lives in the chart-generated secret
  (`kubectl -n monitoring get secret kube-prometheus-stack-grafana -o
  jsonpath='{.data.admin-password}' | base64 -d`); the
  `golden-path-overview` dashboard is provisioned from Git.
- No data usually means: pods lack the scrape annotations (observability not
  enabled in the Service Definition), the `golden-path` scrape job is down,
  or retention (6h) already dropped the window.

## Destroy disposable AWS environment

From repository root, run `make destroy`. The script creates and displays a
destroy plan and requires typing `destroy`. After apply, verify EKS, EC2, NAT,
EIP and CloudWatch resources are absent. This procedure has not been run for
this repository.

## Argo CD failure

Do not delete healthy application workloads. Restore the Argo CD deployment,
verify repository access, then inspect the reconciliation queue and sync status.
Use the guarded `scripts/phase4-experiments.sh argocd-down|argocd-up` actions
only on the named disposable cluster; they save and restore the original
replica count. Application data plane impact and recovery are recorded in
`docs/FAILURE_EXPERIMENTS.md`, and remain NOT RUN until observed.

## Prometheus failure

Do not restart application workloads. Restore Prometheus with
`scripts/phase4-experiments.sh observability-up`, then prove a Ready Pod,
Golden Path targets UP, fresh sample timestamps, a functional SLI query, and
Grafana data. Missing samples during the outage are a blind spot, not evidence
of application downtime.
