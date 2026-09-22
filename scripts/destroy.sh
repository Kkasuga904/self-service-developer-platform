#!/usr/bin/env bash
set -euo pipefail

if [[ ! -f "infra/main.tf" || ! -f "Makefile" ]]; then
  echo "Run this command from the self-service-developer-platform repository root." >&2
  exit 1
fi

terraform -chdir=infra init
terraform -chdir=infra plan -destroy -out=destroy.tfplan
terraform -chdir=infra show destroy.tfplan

if [[ "${AUTO_APPROVE_DESTROY:-false}" != "true" ]]; then
  read -r -p "Apply the reviewed destroy plan? Type 'destroy': " confirmation
  if [[ "${confirmation}" != "destroy" ]]; then
    echo "Destroy cancelled."
    exit 1
  fi
fi

terraform -chdir=infra apply destroy.tfplan
