.PHONY: build test validate render bootstrap destroy

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

destroy:
	bash scripts/destroy.sh
