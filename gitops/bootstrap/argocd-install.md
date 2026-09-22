# Argo CD bootstrap boundary

Argo CD is the one imperative bootstrap dependency. Terraform creates the AWS and
EKS foundation; `scripts/bootstrap-gitops.sh` installs the pinned Argo CD chart
and applies `root-application.yaml`. After that, Argo CD reconciles the platform
and developer services from Git.

Before bootstrap, publish this directory as the root of
`https://github.com/Kkasuga904/self-service-developer-platform.git`, or update
both repository URLs in the bootstrap and ApplicationSet manifests. The example
repository has not been published or reconciled during Phase 2.
