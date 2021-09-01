# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

export GO111MODULE=on

all: clean
	@echo -e "\nBuilding HPA Placement Controller .."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -tags netgo -o ../../bin/hpa-plc/hpa-plc ./cmd/main.go

# The following is done this way as each patch on CI runs build and each merge runs deploy. So for build we don't need to build binary and hence
# no need to create a static binary with additional flags. However, for generating binary, additional build flags are necessary. This if used with
# mock plugin errors out for unit tests. So the seperation avoids the error.

build: clean test cover
deploy: build

.PHONY: test
test: clean
	go test -race ./...

testv: clean
	go test -v -race ./...
	#go test -v -ginkgo.v -race ./...

format:
	@go fmt ./...

clean:
	@rm -f ../../bin/hpa-plc/hpa-plc coverage.html coverage.out

.PHONY: cover
cover:
	@go test -race ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

coverage:
	@go test -race ./... -coverprofile full_cover.out
	@go tool cover -func full_cover.out
	@echo -e "============================================================"
	@echo -e "HPA Placement Total Coverage => " `go tool cover -func full_cover.out | grep total:`
	@echo -e "============================================================"

analyze: lint sec-analyze vet

sec-analyze:
	@echo -e "\n\n Running gosec \n\n"
	@gosec ./...
vet:
	@echo -e "\nRunning go vet .."
	@go vet ./...
lint:
	@echo -e "\nRunning golanfci-lint .."
	@golangci-lint run --enable-all --disable=wsl --disable=lll --disable=gochecknoglobals
	#@golangci-lint run
	#golangci-lint run --no-config --issues-exit-code=0 --deadline=30m \
		--disable-all --enable=deadcode  --enable=gocyclo --enable=golint --enable=varcheck \
		--enable=structcheck --enable=maligned --enable=errcheck --enable=dupl --enable=ineffassign \
		--enable=interfacer --enable=unconvert --enable=goconst --enable=gosec --enable=megacheck
