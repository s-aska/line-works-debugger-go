.DEFAULT_GOAL := help

PROJECT_NAME="line-works-debugger-go"

create: ## GCPプロジェクト作成
	gcloud app create --project $(PROJECT_NAME)

installdeps: ## go mod vendor
	GO111MODULE=on go mod vendor

dev-all: ## dev_appserver.py
	GO111MODULE=on dev_appserver.py \
		`pwd`/app.yaml --support_datastore_emulator=False

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: help
