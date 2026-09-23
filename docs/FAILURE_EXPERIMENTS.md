# Intentional failure experiments

This document is the Phase 4 evidence ledger. It is separate from
[INCIDENTS.md](INCIDENTS.md): that file records unexpected implementation and
AWS incidents, while this file records deliberately injected failures.

Status meanings are the same as [VALIDATION.md](VALIDATION.md): **PASS** means
the observed result matched the stated expectation, **PARTIAL** means some but
not all required evidence exists, **FAIL** means observation contradicted the
hypothesis, and **NOT RUN** means there is no execution evidence. A failed
hypothesis is not rewritten after the fact.

## Execution safety and order

The experiments use only kubectl, Kubernetes primitives, Argo CD, Prometheus,
AWS APIs, Terraform, Git, and shell. No chaos framework is installed. The
required order is:

```text
healthy baseline -> inject one failure -> observe -> recover in source of truth
-> prove baseline restored -> next experiment
```

`scripts/phase4-experiments.sh` refuses mutating actions unless the active
kube context contains `developer-platform-dev`; it saves original replica
counts rather than guessing them during recovery. Raw output belongs in a
temporary local directory, not Git. This ledger keeps timestamps, commands,
small decisive excerpts, masked identifiers, and the conclusion.

## Phase 4 execution record

Attempt started: 2026-09-23 JST. Git revision: `442b36b` before Phase 4 edits.
AWS profile `portfolio` existed, but its SSO session was expired. The initial
`aws sts get-caller-identity --profile portfolio` returned an invalid/expired
SSO-session error. No AWS or Kubernetes mutation had occurred at that point.

> Update this section only from observed output. Never promote a template
> below to PASS based on design inspection or a previous phase.

## Required pre-experiment baseline

Status: **PASS** (2026-09-22 23:43Z)

Command: `AWS_PROFILE=portfolio make phase4-baseline`

Required evidence:

| Signal | Required state | Observed |
| --- | --- | --- |
| EKS | `ACTIVE`, expected version | ACTIVE, 1.35 |
| Nodes | all `Ready` | 2/2 Ready |
| Argo CD Applications | payments/orders `Synced`, `Healthy` | root/payments/orders Synced + Healthy |
| Workloads | payments 3/3, orders 2/2 Ready | 3/3 and 2/2 |
| Health endpoints | both HTTP 200 | both `ok`, HTTP 200 |
| Prometheus | Golden Path targets `up == 1`; SLI query evaluates | 5/5 UP; availability result 1 |
| Kyverno | deployments available; policies ready | all 8 ClusterPolicies Ready |
| RBAC | payments read=yes, cross-team patch=no | yes / no |
| Git revision | Application revision equals intended commit | `46764c0` |

No experiment may start until every row is healthy.

## Experiment A — policy-valid bad deployment

Result: **PARTIAL**

### Hypothesis

Changing `payment-api` to a syntactically valid but nonexistent immutable ECR
digest passes contract, schema, Helm, and Kyverno checks. Argo CD accepts the
Git revision but Kubernetes cannot pull the new image. Because the Deployment
uses `RollingUpdate`, `maxUnavailable: 0`, and `maxSurge: 1`, the three old
Ready replicas should remain behind the Service while one new Pod reports
`ImagePullBackOff`. The Application is expected to be Synced to desired state
but Degraded/Progressing rather than Healthy. Availability should remain near
100%, subject to node pressure and measurement granularity.

### Injection

From a dedicated Phase 4 branch, replace only the final hexadecimal character
of the `payment-api` digest with another valid hexadecimal character. Run the
normal local/CI checks, commit, and push/merge through the same Git path used by
ApplicationSet. Record both SHA and UTC timestamp. This is runtime-invalid but
policy-valid; do not use `latest` or violate an admission rule.

### Expected blast radius

One payment-api rollout and one surge Pod in `team-payments`. The orders data
plane, platform controllers, policy, and monitoring should remain healthy.

### Observed behavior

Bad commit `8dcc9e0` was pushed at 23:45:38Z and passed the contract and Helm
checks. The Application did not automatically refresh within six minutes even
though `timeout.reconciliation` was 180s; a hard refresh at 23:51:18Z advanced
it to the bad SHA at 23:51:27Z. Argo showed `Synced/Progressing`. The Deployment
showed 3 desired, 1 updated, 4 total, 3 available; old ReplicaSet was 3/3 and
the surge ReplicaSet 0/1 Ready. The new Pod reported `ErrImagePull`, then
`ImagePullBackOff`, with ECR digest-not-found Events. Orders remained healthy.

### Detection

Expected signals are Argo health, Deployment unavailable/Progressing
conditions, Pod `ImagePullBackOff`, image-pull Events, ready-replica metric, and
the failed rollout status. Prometheus application metrics may remain healthy
because the old Pods continue serving; this is an important distinction, not
a monitoring defect.

### Developer impact

Expected: deployment of the new revision is blocked, while requests continue
on the old revision. The shortest diagnostic path is Argo Application ->
Deployment condition -> new ReplicaSet/Pod -> Events.

### Platform impact

The contract and policies prevent unsafe structure and privilege, not image
existence or successful startup. GitOps continuously reports desired-state
health but does not infer that the previous Git revision should be restored.

### Data plane impact

The three old Pods remained Ready and behind the Service. A 50-request window
returned 50 HTTP 200 / 0 failures. Prometheus showed the three old targets at
`up=1`, the failed surge target at `up=0`, and available replicas=3.

### Control plane impact

Expected: none. Argo CD and Kubernetes continue reconciling the bad desired
state and expose the failure.

### Recovery

Use `git revert <bad-sha>` (or a correcting commit), push/merge it, and let
Argo CD reconcile. Do not patch the Deployment. Record recovery-start and
Healthy timestamps.

### Recovery evidence

Fix commit `77c4b18` was pushed at 23:53:28Z. Argo automatically observed it
and returned `Synced/Healthy` at 23:56:18Z; recovery took about 170s. Deployment
returned to 3/3 and the failed ReplicaSet was removed from serving.

### Difference from hypothesis

The availability hypothesis was correct, but automatic detection of the bad
commit was not: a hard refresh was required after more than six minutes. Argo
reported the applied bad manifest as `Synced/Progressing`, not Degraded. The
experiment is PARTIAL rather than PASS because normal detection breached the
configured 180s expectation.

### Improvement

Production extensions could include registry/digest existence validation,
deployment-progress alerting, and progressive delivery. Automatic rollback is
intentionally absent: with Git as source of truth, an imperative rollback
would create drift while Git still requests the bad revision.

## Experiment B — Argo CD control-plane failure

Result: **PARTIAL**

### Hypothesis

Scaling only `argocd-application-controller` to zero stops Application
reconciliation, drift detection, and self-heal, but does not stop existing
application Pods or Services. A Git change made during the outage should wait
until recovery. **Argo CD failure should not equal application failure** in
this architecture because Argo CD is in the management path, not the request
path.

### Injection

After a healthy baseline, run `scripts/phase4-experiments.sh argocd-down`.
The original StatefulSet replica count is saved locally. Do not delete Argo
resources, workloads, or namespaces.

### Expected blast radius

GitOps reconciliation and Application status freshness cluster-wide. Existing
payments/orders serving paths should remain available. Any new Git change or
manual drift remains unapplied/unhealed during the outage.

### Observed behavior

The application-controller was scaled 1→0 at 23:56:47Z. Payments stayed 3/3,
Prometheus stayed Running, and 50/50 requests returned HTTP 200. Commit
`07ae85c` was pushed at 23:57:32Z and a harmless Deployment annotation was
added. Argo remained on old SHA `77c4b18`, displayed stale `Synced/Healthy`,
and the annotation remained present: reconciliation and drift detection had
stopped while the data plane continued.

### Detection

Expected: controller replica metric/Pod absence and stale Application status.
Application `/healthz` alone cannot detect this control-plane failure.

### Developer impact

Expected: existing service remains usable, but new deployments and drift
repair stop; Argo status is stale and cannot be trusted as current.

### Platform impact

Loss of deployment convergence, self-heal, and current GitOps visibility.
Kubernetes scheduling and already-created workload resources continue.

### Data plane impact

No measured data-plane outage: 50 successes / 0 failures, 3 Ready replicas.
**Argo CD failure did not equal application failure.**

### Control plane impact

Expected: reconciliation unavailable until the controller returns. API server,
Kyverno, and monitoring are separate controls and should stay operational.

### Recovery

Run `scripts/phase4-experiments.sh argocd-up`; verify the controller Ready,
pending Git change applied, injected drift healed, and every Application
Synced/Healthy. Then rerun the baseline.

### Recovery evidence

Controller Ready returned about 20s after scale-up. Automatic Git/drift pickup
still had not happened after more than three minutes; a hard refresh at
00:02:58Z advanced the Application to `07ae85c`. A managed replicas drift
(3→2) was self-healed to spec=3 immediately after refresh and Ready returned
3/3 by 00:04:35Z. Final cleanup commit `ae78f40` restored the original contact.

### Difference from hypothesis

The main hypothesis held for application traffic. Recovery was slower and less
automatic than predicted. An arbitrary extra Deployment metadata annotation
was not considered drift and was not removed; managed `.spec.replicas` was.
The result is PARTIAL because controller readiness alone did not restore timely
reconciliation without a hard refresh.

### Improvement

For production, run Argo CD components HA, alert on reconciliation age and
controller availability, and define a documented deployment freeze/fallback.
HA reduces probability; it does not place Argo in the application data path.

## Experiment C — observability failure

Result: **PASS**

### Hypothesis

Scaling the Prometheus StatefulSet to zero removes scrape, query, SLI, rule,
and Grafana datasource availability while application health and traffic
continue. Observability failure and application failure are separate states.

### Injection

After restored baseline, run `scripts/phase4-experiments.sh
observability-down`. Grafana and the applications are not stopped.

### Expected blast radius

Prometheus-backed visibility for all teams. Application serving, GitOps,
admission, and Kubernetes scheduling should be unaffected.

### Observed behavior

Directly scaling the Prometheus StatefulSet to zero was first self-healed by
Prometheus Operator within about 34s. The effective injection therefore scaled
the operator 1→0 and Prometheus 1→0 at 00:06:44Z. Prometheus Service endpoints
were empty; queries and SLI calculation were unavailable. Grafana remained
Ready but had no datasource data. Payments remained 3/3, Argo
`Synced/Healthy`, and 50/50 health requests returned 200.

### Detection

Expected: Prometheus Pod absence and failed query/datasource. The current stack
cannot reliably alert through itself while Prometheus is down; Alertmanager is
also intentionally disabled. This is a known blind spot.

### Developer impact

Expected: service works, but historical/current metrics, dashboards, SLI, and
Prometheus-rule visibility are unavailable. Kubernetes status and direct
health checks remain available.

### Platform impact

Loss of telemetry ingestion and evidence during the outage. Missing samples
are not reconstructed after recovery.

### Data plane impact

Observed none: payment-api stayed 3/3 and 50/50 health requests returned 200.

### Control plane impact

The observability control surface is unavailable; GitOps and admission should
remain healthy.

### Recovery

Run `scripts/phase4-experiments.sh observability-up`; then verify the Pod
Ready, `up{job="golden-path"} == 1`, fresh sample timestamps, SLI query result,
and Grafana panels. Rerun baseline.

### Recovery evidence

Operator and Prometheus were restored at 00:09:10Z. Both were Ready at
00:09:34Z (about 24s); by 00:10:27Z all 5 Golden Path targets were `up=1` and
the availability query returned 1 (about 76s from recovery start).

### Difference from hypothesis

The application/monitoring separation matched the hypothesis. The additional
finding was that Prometheus Operator protects the StatefulSet against a direct
scale injection. Stopping Prometheus required pausing its reconciler too.

### Improvement

Production needs independent monitoring of monitoring, alert delivery outside
the failed Prometheus instance, HA/remote durable storage where justified, and
explicit missing-data semantics. Those are documented extensions, not Phase 4
feature additions.

## Experiment D — team boundary / unauthorized action

Result: **PASS**

### Hypothesis

The payments developer group can read payments resources/logs and create a
payments port-forward, but cannot read orders resources, mutate orders
workloads, or create cluster-scoped RBAC. A real server-side dry-run patch
against `order-api` should be rejected by authorization before mutation.

### Injection

Run `scripts/phase4-experiments.sh rbac`. This uses Kubernetes impersonation
from the existing administrator context; it does not claim that an external
IdP login was tested.

### Expected blast radius

None: authorization checks are read-only and the negative patch also uses
server dry-run. Unexpected authorization success is a test failure and the
script exits immediately.

### Observed behavior

At 00:10:58Z, the payments identity returned yes for own Pod read/log and
port-forward (`--subresource=portforward`), no for orders read/patch and
clusterrole create. Real payments Pod list, log read, and port-forward
`/healthz` HTTP 200 succeeded. A real server-dry-run create of an orders
Deployment returned Forbidden (`cannot create deployments.apps`) and no object
was created.

### Detection

The API authorization result and audit trail (if enabled) are the relevant
signals. Prometheus application availability is not evidence of correct RBAC.

### Developer impact

Expected: self-service diagnostics inside the owned namespace; Git changes,
not direct workload mutation, remain the deployment path.

### Platform impact

Expected: ownership boundary enforced without interrupting workloads.

### Data plane impact

Expected none.

### Control plane impact

Expected authorization rejection only.

### Recovery

No mutation is expected. Verify the forbidden annotation is absent and rerun
baseline. If it exists, mark FAIL, remove it through the administrator/GitOps
path, and treat the authorization grant as a blocking defect.

### Recovery evidence

No recovery mutation was needed; the unauthorized Deployment and annotation
were absent and workloads remained healthy.

### Difference from hypothesis

The RBAC hypothesis held. The shorthand `can-i create pods/portforward` gave a
misleading `no`; the correct kubectl form is `create pods
--subresource=portforward`, which returned `yes` and matched the real action.

### Improvement

Production would connect real workforce identities, test EKS access entries
end-to-end, enable/audit authorization logs, and add NetworkPolicy and quotas.
Namespace plus RBAC is an administrative boundary, not hostile-tenant
isolation.

## Optional Experiment E — node maintenance

Result: **NOT RUN (intentionally omitted unless time/cost permit)**

The repository already covers Kubernetes rescheduling mechanics elsewhere.
Run this only if it tests the platform-user view—PDB, replica distribution,
GitOps state, and measured SLI during a planned drain—without reducing to a
Pod deletion exercise. Omission is preferred over duplicate evidence.

## Measurement sheet

Fill one row per experiment from UTC timestamps and fixed-window request
counts. Short disposable-cluster samples are not production benchmarks.

| Experiment | Injected | Detected | Recovery start | Healthy | Detect time | Recover time | HTTP success/fail | Ready before/during/after |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| A | 23:45:38Z | 23:51:27Z (hard refresh) | 23:53:28Z | 23:56:18Z | ~349s | ~170s | 50/0 | 3/3/3 |
| B | 23:56:47Z | immediate controller absence | 23:58:53Z | 00:05:14Z | seconds | ~381s | 50/0 | 3/3/3 |
| C | 00:06:44Z effective | immediate endpoint loss | 00:09:10Z | 00:10:27Z | seconds | ~76s | 50/0 | 3/3/3 |
| D | 00:10:58Z | immediate Forbidden | n/a | 00:11:21Z | seconds | n/a | n/a | unchanged |
