BINARY_NAME = agartha
OS := $(shell uname -s)
src_dir = $(CURDIR)
build_dir = $(CURDIR)/bin
debug_dir = $(build_dir)/debug
release_dir = $(build_dir)/release
.DEFAULT_GOAL := build

COMPILE_DATE :=$(or $(GITHUB_DATE),$(shell date -u +"%Y-%m-%dT%H:%M:%SZ"))
DATE := $(shell TZ=UTC0 git show --quiet --date='format-local:%Y%m%d%H%M%S' --format="%cd" | tail -n 1)
PSEUDODATE := $(or $(GITHUB_PSEUDODATE),$(shell date -u '+%Y%m%d%H%M%S'))
PREFIX = github.com/PaulChristophel/agartha
BUILD := $(or $(SHORT_SHA),$(shell git rev-parse --short=12 HEAD))
FULLBUILD := $(or $(GITHUB_SHA),$(shell git rev-parse HEAD))
VERSION :=  $(or $(GITHUB_VERSION),$(shell cat VERSION))
GOVERSION = $(shell go version)

build:
ifdef ENV
	mkdir -p $(build_dir)/$(ENV)
	go fmt ./...
ifeq ($(ENV), debug)
LDFLAGS=-"-X $(PREFIX)/server/routes.Build=$(FULLBUILD) -X '$(PREFIX)/server/routes.Version=v$(VERSION)-$(DATE)-$(BUILD)' -X '$(PREFIX)/server/routes.GoVersion=$(GOVERSION)' -X '$(PREFIX)/server/routes.CompileDate=$(COMPILE_DATE)' -X '$(PREFIX)/server/routes.CommitDate=$(DATE)'"
build: build-web-dev swagger build-go-dev
else ifeq ($(ENV), release)
LDFLAGS="-w -s -X $(PREFIX)/server/routes.Build=$(FULLBUILD) -X '$(PREFIX)/server/routes.Version=v$(VERSION)' -X '$(PREFIX)/server/routes.GoVersion=$(GOVERSION)' -X '$(PREFIX)/server/routes.CompileDate=$(COMPILE_DATE)' -X '$(PREFIX)/server/routes.CommitDate=$(DATE)'"
build: build-web swagger build-go
endif
else
	$(error ENV not set)
endif

build-web: fmt-web lint-web-fix
	PNPMVERSION="$(VERSION)" pnpm --dir $(src_dir)/web run set:version
	pnpm --dir $(src_dir)/web run build
	pnpm --dir $(src_dir)/web run re:version

build-web-dev: fmt-web lint-web-fix
	PNPMVERSION="$(VERSION)-$(DATE)-$(BUILD)" pnpm --dir $(src_dir)/web run set:version
	pnpm version
	DEBUG=vite:* NODE_ENV=development pnpm --dir $(src_dir)/web run build --mode=development
	pnpm --dir $(src_dir)/web run re:version

build-go: fmt-go lint-go
	@echo "USING LDFLAGS=$(LDFLAGS) FOR BUILD"
	git tag -f v$(VERSION)
	go build -ldflags=$(LDFLAGS) -v -o $(release_dir)/${BINARY_NAME}

build-go-dev: fmt-go lint-go
	@echo "USING LDFLAGS=$(LDFLAGS) FOR BUILD"
	git tag -f v$(VERSION)-$(DATE)-$(BUILD)
	go build -v -ldflags=$(LDFLAGS) -o $(debug_dir)/${BINARY_NAME}

run: watch
watch: watch-go run-web run-compose

watch-go: swagger
	mkdir -p $(debug_dir)
	if [ ! -d "$(src_dir)/web/dist" ]; then 	mkdir -p $(src_dir)/web/dist; touch $(src_dir)/web/dist/migrate; fi
	GIN_MODE=debug air

run-go:
ifeq ($(ENV), release)
	GIN_MODE=release go run main.go serve
else
	GIN_MODE=debug go run main.go serve
endif

run-web:
ifeq ($(ENV), release)
	NODE_ENV=production pnpm --dir $(src_dir)/web dev
else
	NODE_ENV=development pnpm --dir $(src_dir)/web dev --mode=development
endif

preview-web:
ifeq ($(ENV), release)
	NODE_ENV=production pnpm --dir $(src_dir)/web start
else
	NODE_ENV=development pnpm --dir $(src_dir)/web start --mode=development
endif

migrate:
ifeq ($(ENV), release)
	GIN_MODE=release go run main.go migrate
else
	GIN_MODE=debug go run main.go migrate
endif

run-compose: podman-clean
	podman compose -f docker-compose-bare.yaml up

migrate-ci:
	git config --global --add safe.directory /app
	mkdir -p $(src_dir)/web/dist
	touch $(src_dir)/web/dist/migrate
	GIN_MODE=$(ENV) go run main.go migrate
# We need this to keep podman builds from failing "on the first thing that exits"
	# sleep 1200

clean-go:
	go clean
	go mod tidy
	rm -f $(debug_dir)/* $(release_dir)/*

clean-web:
	NODE_ENV=development pnpm --dir $(src_dir)/web rm:all

clean: clean-go clean-web podman-clean
	rm -rf extras/*/__pycache__
	rm -rf .pnpm-store

test: test-go

test-ci: swagger
	git config --global --add safe.directory /app
	if [ ! -d "$(src_dir)/web/dist" ]; then pnpm --dir $(src_dir)/web install; fi
ifeq ($(ENV), release)
	if [ ! -d "$(src_dir)/web/dist" ]; then DEBUG=vite:* NODE_ENV=production pnpm --dir $(src_dir)/web run build --mode=production; fi
	GIN_MODE=$(ENV) go run main.go migrate
else
	if [ ! -d "$(src_dir)/web/dist" ]; then DEBUG=vite:* NODE_ENV=development pnpm --dir $(src_dir)/web run build --mode=development; fi
	GIN_MODE=$(ENV) go run main.go migrate
endif
	go test -v ./...

docker-compose-ci-tests:
	docker-compose -f docker-compose-tests-glibc.yaml up --abort-on-container-exit --exit-code-from test
	docker-compose -f docker-compose-tests-musl.yaml up --abort-on-container-exit --exit-code-from test

test-go:
	if [ ! -d "$(src_dir)/web/dist" ]; then DEBUG=vite:* NODE_ENV=development pnpm --dir $(src_dir)/web run build --mode=development; fi
	go test -v ./...

test-coverage:
	go test ./... -coverprofile=coverage.out

dep-go:
	go get ./...
	go mod download
	go mod verify

configure-go: dep-go

dep-web:
	pnpm --dir $(src_dir)/web install

configure-web: dep-web

dep: dep-web dep-go

configure: configure-web configure-go

upgrade:
	go get -v -u ./...
	pnpm --dir $(src_dir)/web upgrade

vet-go:
	go vet

vet-web:
	pnpm audit --dir $(src_dir)/web

vet: vet-web vet-go

lint-web:
	pnpm --dir $(src_dir)/web lint

lint-web-fix:
	@echo "FIXING LINT ERRORS"
	pnpm --dir $(src_dir)/web lint --fix

lint-go:
	golangci-lint run

lint-go-all:
	golangci-lint run --enable-all

lint: lint-web lint-go

fmt-go:
	go fmt -x ./...

prettier: fmt-web
fmt-web:
	pnpm --dir $(src_dir)/web prettier

fmt: fmt-go fmt-web

reversion:
	pnpm --dir $(src_dir)/web run re:version

swagger:
	@echo "BUILDING API DOCS"
ifeq ($(ENV), release)
	sed -i.bak 's#//	@version		1.0#//	@version		v$(VERSION)#g' main.go
else
	sed -i.bak 's#//	@version		1.0#//	@version		v$(VERSION)-$(DATE)-$(BUILD)#g' main.go
endif
	rm main.go.bak
	swag fmt --exclude $(shell find $(src_dir) -mindepth 1 -maxdepth 1 -type d | grep -v 'server' | tr '\n' ',')
	swag init -p snakecase --generatedTime --pd --pdl 3 --exclude $(shell find $(src_dir) -mindepth 1 -maxdepth 1 -type d | grep -v 'server' | tr '\n' ',') --output $(src_dir)/server/docs/v1
	git config --global --add safe.directory /app
	git restore main.go
	
podman-test: clean
	@echo "GITHUB_DATE=$(DATE)"
	@echo "GITHUB_PSEUDODATE=$(PSEUDODATE)"
	@echo "GITHUB_VERSION=$(VERSION)-$(DATE)-$(BUILD)"
	podman pull docker.io/library/golang:alpine
	podman pull docker.io/library/golang:bookworm
	podman pull docker.io/library/alpine:edge
	podman pull docker.io/library/debian:bookworm
	podman build -f Dockerfile --build-arg=ENV=debug --build-arg=GITHUB_DATE=${DATE} --build-arg=GITHUB_PSEUDODATE=${PSEUDODATE} --build-arg=GITHUB_VERSION=${VERSION}-${DATE}-${BUILD} . -t oitacr.azurecr.io/pmartin47/${BINARY_NAME}:slim-test --target slim
	podman build -f Dockerfile --build-arg=ENV=debug --build-arg=GITHUB_DATE=${DATE} --build-arg=GITHUB_PSEUDODATE=${PSEUDODATE} --build-arg=GITHUB_VERSION=${VERSION}-${DATE}-${BUILD} . -t oitacr.azurecr.io/pmartin47/${BINARY_NAME}:alpine-test --target alpine
	podman build -f Dockerfile-glibc --build-arg=ENV=debug --build-arg=GITHUB_DATE=${DATE} --build-arg=GITHUB_PSEUDODATE=${PSEUDODATE} --build-arg=GITHUB_VERSION=${VERSION}-${DATE}-${BUILD} . -t oitacr.azurecr.io/pmartin47/${BINARY_NAME}:slim-glibc-test --target slim-glibc
	podman build -f Dockerfile-glibc --build-arg=ENV=debug --build-arg=GITHUB_DATE=${DATE} --build-arg=GITHUB_PSEUDODATE=${PSEUDODATE} --build-arg=GITHUB_VERSION=${VERSION}-${DATE}-${BUILD} . -t oitacr.azurecr.io/pmartin47/${BINARY_NAME}:debian-test --target debian

podman-compose: podman-clean
	podman compose -f docker-compose-tests-glibc.yaml up

podman-dev: podman-clean
	podman compose -f docker-compose-dev.yaml up

podman-clean:
	podman compose -f docker-compose-dev.yaml rm -f -v -s
	podman compose -f docker-compose-tests-glibc.yaml rm -f -v -s
	podman compose -f docker-compose-tests-musl.yaml rm -f -v -s
	podman volume prune -f