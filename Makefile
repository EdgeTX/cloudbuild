.PHONY: test lint lint-fix build-server build-worker edgetx-build edgetx-build-install

test:
	go test -v ./...

migrate:
	go run cmd/db/main.go -migrate

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2
	golangci-lint run ./...

lint-fix:
	go mod tidy
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2
	golangci-lint run --fix ./...

build:
	rm -rf ./bin/ebuild
	go build -o ./bin/ebuild -trimpath ./cmd/ebuild

edgetx-build:
	rm -rf ./bin/edgetx-build
	go build -o ./bin/edgetx-build -trimpath ./cmd/edgetx-build/main.go

edgetx-build-install:
	go install -trimpath ./cmd/edgetx-build/main.go
