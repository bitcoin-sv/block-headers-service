SHELL=/bin/bash

run-grpc-server:
	@go run -race cmd/grpc-server/main.go

run-all-tests: run-linter run-unit-tests

pre-commit: vendor-deps run-all-tests

run-unit-tests:
	@go clean -testcache && go test -v ./... -race

run-pipeline-unit-tests:
	@go clean -testcache && go test -v ./... -race -tags pipeline

run-unit-tests-cover:
	@go test ./... -race -v -coverprofile cover.out && \
	go tool cover -html=cover.out -o cover.html && \
	open file:///$(shell pwd)/cover.html

run-linter:
	@golangci-lint run --deadline=240s --skip-dirs=vendor --tests

install-linter:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.35.2

go-doc-mac:
	@open http://localhost:6060 && \
	godoc -http=:6060

go-doc-linux:
	@xdg-open http://localhost:6060 && \
	godoc -http=:6060

build-image:
	@docker-compose -f docker-compose.build.yml build --no-cache

run-compose:
	@docker-compose up

run-compose-d:
	@docker-compose up -d

fresh-compose:
	@docker-compose down && docker-compose build && docker-compose up

stop-compose:
	@docker-compose down

run-compose-mysql:
	@docker-compose -f docker-compose.yml -f docker-compose.mysql.yml up

compose-local-mysql:
	@docker-compose -f docker-compose.yml -f docker-compose.local.yml -f docker-compose.mysql.yml up

compose-local-postgres:
	@docker-compose -f docker-compose.yml -f docker-compose.local.yml -f docker-compose.postgres.yml up

run-compose-postgres:
	@docker-compose -f docker-compose.yml -f docker-compose.postgres.yml up

vendor-deps:
	@go mod tidy && go mod vendor
