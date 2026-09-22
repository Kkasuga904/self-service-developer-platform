.PHONY: build test validate render bootstrap phase4-baseline destroy

build:
	go build -o platform ./cmd/platform

test:
	go test ./...
	cd sample-app && go test ./...

render:
	helm template payment-api charts/golden-path -f services/payments-team/payment-api/service.yaml

validate:
	bash scripts/local-validate.sh

bootstrap:
	bash scripts/bootstrap-gitops.sh

phase4-baseline:
	bash scripts/phase4-experiments.sh baseline

destroy:
	bash scripts/destroy.sh
