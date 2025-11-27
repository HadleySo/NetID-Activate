GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
VERSION=0.2.2
HASH ?= $(shell git rev-parse HEAD)
BINARY_NAME=NetIdActivate_$(VERSION)_$(GOOS)_$(GOARCH)

build:
	go build -ldflags "-X github.com/hadleyso/netid-activate/src/routes.Version=$(VERSION) -X github.com/hadleyso/netid-activate/src/routes.GitCommit=$(HASH)" -o ./bin/${BINARY_NAME} app.go

run:
	go build -ldflags "-X github.com/hadleyso/netid-activate/src/routes.Version=$(VERSION) -X github.com/hadleyso/netid-activate/src/routes.GitCommit=$(HASH)" -o ./bin/${BINARY_NAME} app.go
	./bin/${BINARY_NAME}

clean:
	go clean
	rm ./bin/${BINARY_NAME}

test:
	go test -cover ./src/auth
	go test -cover ./src/attribute
	go test -cover ./src/db
	go test -cover ./src/handlers
	go test -cover ./src/redhat-idm
	
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
