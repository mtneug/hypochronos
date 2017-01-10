# Copyright (c) 2016 Matthias Neugebauer <mtneug@mailbox.org>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GIT_COMMIT=$(shell git rev-parse --short HEAD || echo "unknown")
GIT_TREE_STATE=$(shell sh -c 'if test -z "`git status --porcelain 2>/dev/null`"; then echo clean; else echo dirty; fi')
BUILD_DATE=$(shell date -u +"%Y-%m-%d %T %Z")

CMD=hypochronosd hypochronos-json-example
BIN=$(addprefix bin/, $(CMD))

PKG=$(shell cat .godir)
PKG_INTEGRATION=${PKG}/integration
PKGS=$(shell go list ./... | grep -v /vendor/)

GO_LDFLAGS=-ldflags " \
	-s \
	-X '$(PKG)/version.gitCommit=$(GIT_COMMIT)' \
	-X '$(PKG)/version.gitTreeState=$(GIT_TREE_STATE)' \
	-X '$(PKG)/version.buildDate=$(BUILD_DATE)'"
GO_BUILD_ARGS=-v $(GO_LDFLAGS)

GOMETALINTER_COMMON_ARGS=\
	--sort=path \
	--vendor \
	--tests \
	--vendored-linters \
	--disable-all \
	--enable=gofmt \
	--enable=vet \
	--enable=vetshadow \
	--enable=golint \
	--enable=ineffassign \
	--enable=goconst \
	--enable=goimports \
	--enable=staticcheck \
	--enable=unused \
	--enable=misspell \
	--enable=lll \
	--line-length=120

all: lint build test integration
ci: lint-full build coverage coverage-integration

build: $(BIN)
	@echo "⌛ $@"

bin/%: cmd/% FORCE
	@echo "⌛ $@"
	@go build $(GO_BUILD_ARGS) -o $@ ./$<

install: $(addprefix install-, $(CMD))
		@echo "⌛ $@"

install-%: cmd/% FORCE
	@echo "⌛ $@"
	@go install $(GO_BUILD_ARGS) ./$<

run: bin/hypochronosd
	@echo "⌛ $@"
	@bin/hypochronosd \
		--log-level debug \
		--service-update-period 1s \
		--node-update-period 1s \
		--default-period 1m \
		--default-state deactivated \
		--default-minimum-scheduling-duration 10s

clean:
	@echo "⌛ $@"
	@rm -f bin

generate:
	@echo "⌛ $@"
	@go generate -v -x ${PKGS}

lint:
	@echo "⌛ $@"
	@test -z "$$(gometalinter --deadline=10s ${GOMETALINTER_COMMON_ARGS} ./... | grep -v '.pb.go:' | tee /dev/stderr)"

lint-full:
	@echo "⌛ $@"
	@test -z "$$(gometalinter --deadline=5m ${GOMETALINTER_COMMON_ARGS} \
			--enable=deadcode \
			--enable=varcheck \
			--enable=structcheck \
			--enable=errcheck \
			--enable=unconvert \
			./... \
		| grep -v '.pb.go:' \
		| tee /dev/stderr)"

test:
	@echo "⌛ $@"
	@go test -parallel 8 -race $(filter-out ${PKG_INTEGRATION},${PKGS})

integration:
	@echo "⌛ $@"
	@go test -parallel 8 -race ${PKG_INTEGRATION}

coverage:
	@echo "⌛ $@"
	@status=0; \
	for pkg in $(filter-out ${PKG_INTEGRATION},${PKGS}); do \
		go test -race -coverprofile="../../../$$pkg/coverage.txt" -covermode=atomic $$pkg; \
		true $$((status=status+$$?)); \
	done; \
	exit $$status

coverage-integration:
	@echo "⌛ $@"
	@go test -race -coverprofile="../../../${PKG_INTEGRATION}/coverage.txt" -covermode=atomic ${PKG_INTEGRATION}

FORCE:

.PHONY: all ci build clean generate lint lint-full test integration coverage coverage-integration FORCE
