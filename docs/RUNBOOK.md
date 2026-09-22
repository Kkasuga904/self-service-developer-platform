# Runbook

## Local contract failure

Run `./platform validate <path>`. Correct every reported field; do not bypass
the validator by editing rendered manifests because Argo CD renders from the
contract again.

## Argo Application is OutOfSync

This procedure is unvalidated until EKS testing:

1. Inspect `argocd app get <service>` and related Kubernetes events.
2. Compare the Application source revision with main.
3. Render the exact values locally with Helm.
4. Fix desired state in Git; avoid an imperative patch.
5. Sync and record the revision and observed recovery.

## Destroy disposable AWS environment

From repository root, run `make destroy`. The script creates and displays a
destroy plan and requires typing `destroy`. After apply, verify EKS, EC2, NAT,
EIP and CloudWatch resources are absent. This procedure has not been run for
this repository.

## Argo CD failure

Do not delete healthy application workloads. Restore the Argo CD deployment,
verify repository access, then inspect the reconciliation queue and sync status.
Application data plane impact and recovery remain Phase 4 validation subjects.
