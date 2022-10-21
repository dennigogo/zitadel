COMMIT := $(shell git rev-parse --short HEAD)
ifeq ($(shell git tag --contains HEAD),)
  VERSION := $(shell git rev-parse --short HEAD)
else
  VERSION := $(shell git tag --contains HEAD)
endif
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOLDFLAGS += -X github.com/zitadel/dennigogo/cmd/build.version=$(VERSION)
GOLDFLAGS += -X github.com/zitadel/dennigogo/cmd/build.commit=$(COMMIT)
GOLDFLAGS += -X github.com/zitadel/dennigogo/cmd/build.date=$(BUILDTIME)

GOFLAGS = -ldflags "$(GOLDFLAGS)"
NAME := zitadel

.PHONY: build base tools console zitadel

build: ## Build application
	GOSUMDB=off \
	go build $(GOFLAGS) -o $(NAME)

tools:
	./tools/install.sh

base: tools
	docker build -f build/grpc/Dockerfile -t zitadel-base:local .

console: base
	rm openapi/statik/statik.go || true
	docker build --no-cache -f build/zitadel/Dockerfile . -t zitadel-go-test --target go-codecov -o .artifacts/codecov
	docker build -f build/zitadel/Dockerfile . -t zitadel-go-base --target go-copy
	docker run -t -v ${PWD}/.artifacts/grpc/go-client:/copy_internal  zitadel-go-base /bin/sh -c "cp -r /internal/* /copy_internal"
	cp -r .artifacts/grpc/go-client/* internal
	docker run -t -v ${PWD}/.artifacts/pkg:/copy_pkg  zitadel-go-base /bin/sh -c "cp -r /pkg/* /copy_pkg"
	cp -r .artifacts/pkg/* pkg
	docker run -t -v ${PWD}/.artifacts/openapi:/copy_openapi  zitadel-go-base /bin/sh -c "cp -r /openapi/* /copy_openapi"
	cp -r .artifacts/openapi/* openapi
	docker build -f build/console/Dockerfile . -t zitadel-npm-console --target angular-build
	docker run -t -v ${PWD}/.artifacts/grpc/js-client:/copy_generated  zitadel-npm-console /bin/sh -c "cp -r /console/src/app/proto/generated/* /copy_generated"
	docker run -t -v ${PWD}/.artifacts/console:/copy_console  zitadel-npm-console /bin/sh -c "cp -r dist/console/* /copy_console"
	cp -r .artifacts/console/* internal/api/ui/console/static/

zitadel: console
	docker build -f build/Dockerfile . -t zitadel-final:local --target final