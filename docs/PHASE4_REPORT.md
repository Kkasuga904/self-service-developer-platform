# Phase 4 final report

This report is deliberately evidence-gated. Phase 4 implementation artifacts
exist, but real AWS experiments remain **NOT RUN** until fresh output is added
to [FAILURE_EXPERIMENTS.md](FAILURE_EXPERIMENTS.md). This file must not be
rewritten as a success narrative merely because the expected behavior is
plausible.

## Phase 4 Summary

Prepared four bounded experiments: policy-valid bad deployment, Argo CD
application-controller loss, Prometheus loss, and cross-team unauthorized
action. A guarded runner enforces the expected kube context, saves replica
counts for recovery, and keeps experiments sequential. Node maintenance is
optional and intentionally omitted unless it adds platform-level evidence.

## Experiment Results

| Experiment | Result | Evidence |
| --- | --- | --- |
| A — bad deployment | NOT RUN | execution ledger incomplete |
| B — Argo CD control plane | NOT RUN | execution ledger incomplete |
| C — observability | NOT RUN | execution ledger incomplete |
| D — team boundary | NOT RUN | execution ledger incomplete |
| E — node maintenance | NOT RUN / optional | avoid duplicating Pod-rescheduling mechanics |

## Most Important Findings

1. The Golden Path prevents structural and security violations, not a
   policy-valid image pull or startup failure.
2. Application serving, GitOps reconciliation, and observability are separate
   failure domains by design; Phase 4 must measure rather than assume that
   separation.
3. Automatic rollback is intentionally absent because Git remains desired
   state; recovery is a revert/fix commit followed by reconciliation.
4. Namespace RBAC is an administrative ownership boundary, not complete tenant
   isolation.
5. Monitoring loss creates unknown periods; restored Prometheus cannot recreate
   samples never collected.

Items 1–5 are design findings until their corresponding experiment is marked
PASS/PARTIAL/FAIL from observation.

## Hypothesis vs Reality

No comparison is yet supported. The hypotheses and required observations are
written before injection in FAILURE_EXPERIMENTS. Preserve any mismatch.

## Recovery

- A: revert/fix Git commit; wait for Argo and Deployment Healthy.
- B: restore the saved Argo controller replica count; prove pending reconcile
  and drift self-heal.
- C: restore the saved Prometheus replica count; prove fresh targets, samples,
  queries, and dashboards.
- D: no mutation should occur; verify the forbidden annotation is absent.

These are procedures, not completed recovery evidence.

## Platform Guarantees

Supported by Phase 3 evidence: typed contract validation, standard secure
workload rendering, layered Kyverno enforcement, GitOps drift reconciliation,
team-scoped diagnostic RBAC, application metrics/SLI queries, short-lived OIDC,
and disposable AWS teardown. Phase 4 guarantees must wait for execution.

## Platform Non-Guarantees

Successful application startup, image existence, automatic semantic rollback,
hostile tenant isolation, uninterrupted deployment while Argo is unavailable,
visibility while Prometheus is unavailable, external paging, long-window SLO
compliance, multi-cluster/region resilience, or production scale.

## Unexpected Incidents

The Phase 4 start found an expired `portfolio` AWS SSO session. No cloud
mutation occurred before reauthentication. Earlier unplanned incidents remain
preserved in INCIDENTS and VALIDATION, separate from intentional injections.

## Changes Made

- Added the guarded, recovery-aware Phase 4 experiment runner.
- Added a consistent experiment/evidence ledger and final report.
- Added a 20-question interview guide tied to repository evidence.
- Reframed README around problem, design, evidence, failures, and limits.
- Clarified the separation between unexpected incidents and deliberate faults.

## Remaining Limitations

Real experiment output, timestamps, traffic counts, recovery times, fresh cost,
destroy, residual checks, and post-destroy local validation are still required.

## Future Production Extensions

Remote Terraform state/locking, NetworkPolicy, ResourceQuota, image signing,
workforce identity, HA controllers, independent alert routing, durable metrics,
staged policy rollout, longer SLO evaluation, and defined DR objectives. These
are documented only, not a proposed Phase 5.

## AWS Cost

NOT MEASURED for Phase 4. Do not reuse Phase 3 duration as a Phase 4 charge.
Record actual lifetime and available billing evidence after destroy; otherwise
list the charged resource classes without inventing a total.

## Destroy

**NOT RUN** for Phase 4.

## Residual Resource Check

**NOT RUN** for Phase 4.

## Final Local Validation

**NOT RUN** after Phase 4 destroy. Interim local results, if any, belong in
VALIDATION and do not satisfy the required final post-destroy run.

## Final portfolio review

### Strongest evidence

1. Complete real EKS lifecycle with API-verified cleanup in Phase 3.
2. One policy definition exercised in CI and live admission.
3. Observed Argo CD drift/self-heal from Git-managed desired state.
4. Positive and negative GitHub OIDC and team-RBAC boundary checks.
5. Unexpected integration failures retained with diagnosis and fixes.

### Weakest areas

1. Short-lived personal validation is not sustained production operation.
2. Two teams and one cluster do not demonstrate organizational or fleet scale.
3. Namespace/RBAC lacks NetworkPolicy, quotas, and hostile-tenant isolation.
4. Monitoring has no independent alert delivery or long-window SLO evidence.
5. Local Terraform state and single-instance controllers are unsuitable for a
   real shared platform.

### Questions I would ask this candidate

1. Which exact failure classes does the contract prevent?
2. Why should a bad image commit not trigger an imperative rollback?
3. How did old replicas behave during the failed rollout?
4. Does Argo CD failure equal application failure, and what proves it?
5. Which status becomes stale when the application controller is down?
6. What telemetry is permanently lost during Prometheus downtime?
7. Who monitors the monitoring system in this design?
8. Why can payments port-forward but not patch its own Deployment?
9. Why is a namespace not a security boundary?
10. How did port 9443 cause Kyverno admission failure?
11. What prevents the GitHub role from being assumed by another branch/repo?
12. Where could ApplicationSet expand blast radius?
13. Which validations are structural versus observed in EKS?
14. What three changes are blocking before production use?
15. What would change first at 50 teams?

### Claims the candidate can safely make

The candidate built and tested a personal platform prototype, ran the specific
real-AWS behaviors marked PASS in VALIDATION, diagnosed recorded cloud
integration failures, and understands the stated boundaries and trade-offs.

### Claims the candidate should NOT make

Do not claim production operation experience, enterprise readiness,
battle-testing, hostile multi-tenancy, automatic rollback, 30-day SLO
compliance, many-team scale, or Phase 4 results until their evidence rows are
filled from a fresh run.

