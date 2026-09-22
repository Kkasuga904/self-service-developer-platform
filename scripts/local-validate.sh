#!/usr/bin/env bash
set -euo pipefail

go test ./...
go vet ./...
(
  cd sample-app
  go test ./...
  go vet ./...
)
go build -o platform ./cmd/platform
./platform validate services/payments-team/payment-api/service.yaml

terraform -chdir=infra fmt -check -recursive
terraform -chdir=infra init -backend=false
terraform -chdir=infra validate

helm lint charts/golden-path -f services/payments-team/payment-api/service.yaml
helm template payment-api charts/golden-path \
  -f services/payments-team/payment-api/service.yaml > rendered.yaml
helm template payment-api-autoscaled charts/golden-path \
  -f services/payments-team/payment-api/service.yaml \
  --set spec.autoscaling.enabled=true \
  --set spec.autoscaling.minReplicas=2 \
  --set spec.autoscaling.maxReplicas=4 \
  --set spec.autoscaling.targetCPUUtilization=60 > rendered-hpa.yaml

if command -v kubeconform >/dev/null 2>&1; then
  kubeconform -strict -summary \
    -kubernetes-version 1.35.0 \
    rendered.yaml rendered-hpa.yaml
else
  echo "NOT RUN kubeconform: executable not found" >&2
fi

kubectl kustomize gitops/platform >/dev/null
bash -n scripts/bootstrap-gitops.sh scripts/destroy.sh scripts/local-validate.sh

if command -v trivy >/dev/null 2>&1; then
  trivy fs --scanners vuln,secret,misconfig --severity HIGH,CRITICAL \
    --exit-code 1 --skip-dirs infra/.terraform .
else
  echo "NOT RUN trivy: executable not found" >&2
fi

if command -v docker >/dev/null 2>&1; then
  docker build -t platform-sample-app:local sample-app
else
  echo "NOT RUN docker build: executable not found" >&2
fi
