# Policies

Phase 3 implements the enforcement staging decided in ADR 0005: the same
Kyverno ClusterPolicy files in `policies/kyverno/` are used by two layers.

## Two layers, one source

| Layer | Mechanism | Responsibility |
| --- | --- | --- |
| Layer 1 — CI | `kyverno test policies/tests/` in the `policy` CI job | **Fast feedback.** Fails the PR before merge with the same messages admission would produce. |
| Layer 2 — Admission | Kyverno (chart 3.9.1 / app v1.19.1) with the same files applied, `validationFailureAction: Enforce` | **Final enforcement boundary.** Rejects direct `kubectl apply` bypasses of CI and the Golden Path. |

Because both layers consume the identical files, a rule cannot drift between
"what CI previews" and "what the cluster enforces". The install path is
`scripts/bootstrap-kyverno.sh`, which also asserts the installed app version.

## Guardrails (8 ClusterPolicies)

- `disallow-host-namespaces`: no `hostNetwork`/`hostPID`/`hostIPC`.
- `disallow-privileged`: no `privileged`, `allowPrivilegeEscalation` must be `false`.
- `require-run-as-non-root`: pod or container level.
- `restrict-capabilities`: `capabilities.drop` must be `[ALL]`.
- `require-resources`: CPU/memory requests and limits on every container.
- `disallow-latest-tag`: no `:latest`, no untagged images; digests pass.
- `require-owner-labels`: `platform.example.io/owner`, `platform.example.io/service`,
  `platform.example.io/environment`, `app.kubernetes.io/managed-by`.
- `require-probes`: readiness and liveness probes on every container.

All policies match `Pod` CREATE/UPDATE with `background: false` (no audit
noise on the disposable cluster). In-cluster, Kyverno autogen extends Pod
rules to controllers, so the Deployments Argo CD creates are covered without
duplicating JMESPath for `spec.template.spec`.

## Negative tests (cases A–E)

Fixtures and expectations live in `policies/tests/` (one directory per case,
the documented Kyverno layout):

- A `nginx:latest` → REJECT (`disallow-latest-tag`);
- B `privileged: true` → REJECT (`disallow-privileged`);
- C missing requests/limits → REJECT (`require-resources`);
- D missing ownership labels → REJECT (`require-owner-labels`);
- E valid Golden Path-style workload → PASS all policies.

Each case also asserts PASS on the other seven policies to catch
over-blocking. The rendered Deployments for both teams are additionally
asserted in Go (`internal/goldenpath`): ownership labels, probes, resources,
non-root, dropped capabilities, pinned images.

## Exemptions (explicit, not silent)

Platform namespaces are excluded from admission:
`kube-system`, `kube-public`, `kube-node-lease`, `argocd`, `kyverno`,
`monitoring`. Rationale: those components are platform-operated, installed
outside the developer path, and several (e.g. node-level agents) cannot meet
workload guardrails. Exemptions are namespace-scoped, never workload-scoped.

## Known limitations

- Rules cover `containers` only; `initContainers`/`ephemeralContainers` are
  out of scope (documented, Phase 4 if a workload needs them).
- `ClusterPolicy` (legacy Kyverno) is deprecated upstream in v1.19 in favor
  of CEL `ValidatingPolicy`. It is used here because autogen to Deployments
  works without extra plumbing for a small rule set; migration to
  `ValidatingPolicy` is the recorded follow-up (see TRADE_OFFS.md).
- No image signature verification and no vulnerability admission: out of
  scope for this phase by design.
