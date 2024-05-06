BINARY_NAME = agartha
OS := $(shell uname -s)
src_dir = $(CURDIR)
build_dir = $(CURDIR)/bin
debug_dir = $(build_dir)/debug
release_dir = $(build_dir)/release
.DEFAULT_GOAL := build

build:
ifdef ENV
	mkdir -p $(debug_dir) $(release_dir)
	go fmt ./...
ifeq ($(ENV), DEVEL)
build: build-web-dev swagger build-go-dev
else ifeq ($(ENV), PRODUCTION)
build: build-web swagger build-go
endif
else
	$(error ENV not set)
endif

build-web:
	pnpm --dir $(src_dir)/web run build

build-web-dev:
	NODE_ENV=development pnpm --dir $(src_dir)/web run build -- --mode=development

build-go:
	go build -ldflags="-w -s" -o $(release_dir)/${BINARY_NAME}

build-go-dev:
	go build -o $(debug_dir)/${BINARY_NAME}

run:
ifeq ($(ENV), DEVEL)
	GIN_MODE=debug go run main.go serve
else
	GIN_MODE=release go run main.go serve
endif

run-web: build
	NODE_ENV=development pnpm --dir $(src_dir)/web dev

watch:
	mkdir -p $(debug_dir)
	go get -u github.com/cosmtrek/air
	GIN_MODE=debug go run github.com/cosmtrek/air

go-clean:
	go clean
	go mod tidy
	rm -f $(debug_dir)/* $(release_dir)/*

web-clean:
	rm -rf web/node_modules web/package-lock.json web/pnpm.lock web/dist

clean: go-clean web-clean

test:
	go test ./...

test-coverage:
	go test ./... -coverprofile=coverage.out

go-dep:
	go get ./...
	go mod download
	go mod verify

go-configure: go-dep

web-dep:
	pnpm --dir $(src_dir)/web install

web-configure: web-dep

dep: web-dep go-dep

configure: web-configure go-configure

upgrade:
	go get -u ./...

go-vet:
	go vet

web-vet:
	pnpm audit --dir $(src_dir)/web

vet: web-vet go-vet

lint:
	golangci-lint run --enable-all

swagger:
	/Users/pmartin47/go/bin/swag fmt --exclude $(shell find $(src_dir) -mindepth 1 -maxdepth 1 -type d | grep -v 'server' | tr '\n' ',')
	/Users/pmartin47/go/bin/swag init --generatedTime --pd --pdl 3 --exclude $(shell find $(src_dir) -mindepth 1 -maxdepth 1 -type d | grep -v 'server' | tr '\n' ',') --output $(src_dir)/server/docs/v1
	
podman: clean configure
	podman pull docker.io/library/golang:alpine
	podman pull docker.io/library/busybox:latest
	podman pull docker.io/library/alpine:edge
	podman build . -t oitacr.azurecr.io/pmartin47/${BINARY_NAME}:latest --target slim
	podman build . -t oitacr.azurecr.io/pmartin47/${BINARY_NAME}:latest-busybox --target busybox
	podman build . -t oitacr.azurecr.io/pmartin47/${BINARY_NAME}:latest-alpine --target alpine

podman-test: clean configure
	podman pull docker.io/library/golang:alpine
	podman pull docker.io/library/busybox:latest
	podman pull docker.io/library/alpine:edge
	podman build . -t oitacr.azurecr.io/pmartin47/${BINARY_NAME}:latest-test --target slim
	podman build . -t oitacr.azurecr.io/pmartin47/${BINARY_NAME}:latest-busybox-test --target busybox
	podman build . -t oitacr.azurecr.io/pmartin47/${BINARY_NAME}:latest-alpine-test --target alpine