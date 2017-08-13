GO	?= GO15VENDOREXPERIMENT=1 go
GOPATH	:= $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))

PROMU		?= $(GOPATH)/bin/promu
STATICCHECK	?= $(GOPATH)/bin/staticcheck
pkgs		= $(shell $(GO) list ./... | grep -v /vendor/)

PREFIX	?= $(shell pwd)
BIN_DIR	?= $(shell pwd)

all: format vet staticcheck build test

style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

test:
	@echo ">> running tests"
	@$(GO) test -short -race $(pkgs)

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

staticcheck: $(STATICCHECK)
	@echo ">> running staticcheck"
	@$(STATICCHECK) $(pkgs)

build: $(PROMU)
	@echo ">> building binaries"
	@$(PROMU) build --prefix $(PREFIX)

tarball: $(PROMU)
	@echo ">> building release tarball"
	@$(PROMU) tarball --prefix $(PREFIX) $(BIN_DIR)

$(GOPATH)/bin/promu promu:
	@GOOS= GOARCH= $(GO) get -u github.com/prometheus/promu

$(GOPATH)/bin/staticcheck:
	@GOOS= GOARCH= $(GO) get -u honnef.co/go/tools/cmd/staticcheck

.PHONY: all style format build test test-e2e vet tarball docker promu staticcheck

# Declaring the binaries at their default locations as PHONY targets is a hack
# to ensure the latest version is downloaded on every make execution.
# If this is not desired, copy/symlink these binaries to a different path and
# set the respective environment variables.
.PHONY: $(GOPATH)/bin/promu $(GOPATH)/bin/staticcheck
