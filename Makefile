GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
VERSION=0.1.0
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