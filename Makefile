.PHONY: test lint lint-fix build-server build-worker edgetx-build edgetx-build-install

test:
	go test -v ./...

migrate:
	go run cmd/db/main.go -migrate

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1
	golangci-lint run ./...

lint-fix:
	go mod tidy
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1
	golangci-lint run --fix ./...

build-server:
	rm -rf ./bin/server
	go build -o ./bin/server -trimpath ./cmd/server/main.go

build-worker:
	rm -rf ./bin/worker
	go build -o ./bin/worker -trimpath ./cmd/worker/main.go

edgetx-build:
	rm -rf ./bin/edgetx-build
	go build -o ./bin/edgetx-build -trimpath ./cmd/edgetx-build/main.go

edgetx-build-install:
	go install -trimpath ./cmd/edgetx-build/main.go
