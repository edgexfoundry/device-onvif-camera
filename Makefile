.PHONY: build docker test clean prepare update

#GOOS=linux

GO=CGO_ENABLED=0 GO111MODULE=on go
GOCGO=CGO_ENABLED=1 GO111MODULE=on go

# see https://shibumi.dev/posts/hardening-executables
CGO_CPPFLAGS="-D_FORTIFY_SOURCE=2"
CGO_CFLAGS="-O2 -pipe -fno-plt"
CGO_CXXFLAGS="-O2 -pipe -fno-plt"
CGO_LDFLAGS="-Wl,-O1,–sort-common,–as-needed,-z,relro,-z,now"

ARCH=$(shell uname -m)

MICROSERVICES=cmd/device-onvif-camera
.PHONY: $(MICROSERVICES)

# VERSION file is not needed for local development, In the CI/CD pipeline, a temporary VERSION file is written
VERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)

# This pulls the version of the SDK from the go.mod file
SDKVERSION=$(shell cat ./go.mod | grep 'github.com/edgexfoundry/device-sdk-go/v2 v' | awk '{print $$2}')

GIT_SHA=$(shell git rev-parse HEAD)
CGOFLAGS=-ldflags "-linkmode=external \
                   -X github.com/edgexfoundry/device-onvif-camera.Version=$(VERSION) \
                   -X github.com/edgexfoundry/device-sdk-go/v2/internal/common.SDKVersion=$(SDKVERSION)" \
                   -trimpath -mod=readonly -buildmode=pie

build: $(MICROSERVICES)

run: build
	cd cmd && \
		EDGEX_SECURITY_SECRET_STORE=false \
		./device-onvif-camera -c ./res -cp=consul.http://localhost:8500


cmd/device-onvif-camera:
	$(GOCGO) build $(CGOFLAGS) -o $@ ./cmd

docker:
	docker build . \
		--build-arg http_proxy=$(HTTP_PROXY) \
		--build-arg https_proxy=$(HTTPS_PROXY) \
		--build-arg no_proxy=$(NO_PROXY) \
		-t edgexfoundry/device-onvif-camera:$(GIT_SHA) \
		-t edgexfoundry/device-onvif-camera:$(VERSION)-dev

tidy:
	go mod tidy

unittest:
	$(GOCGO) test ./... -coverprofile=coverage.out ./...

lint:
	@which golangci-lint >/dev/null || echo "WARNING: go linter not installed. To install, run\n  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.42.1"
	@if [ "z${ARCH}" = "zx86_64" ] && which golangci-lint >/dev/null ; then golangci-lint run --config .golangci.yml ; else echo "WARNING: Linting skipped (not on x86_64 or linter not installed)"; fi

test: unittest lint
	$(GOCGO) vet ./...
	gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")
	[ "`gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")`" = "" ]
	./bin/test-attribution-txt.sh

coveragehtml:
	go tool cover -html=coverage.out -o coverage.html

format:
	gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")
	[ "`gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")`" = "" ]

update:
	$(GO) mod download

clean:
	rm -f $(MICROSERVICES)

vendor:
	$(GO) mod vendor
