OUT_DIR := out
PROG := docker-machine-driver-linode

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

ifeq ($(GOOS),windows)
	BIN_SUFFIX := ".exe"
endif

.PHONY: build
build:
	go build -o $(OUT_DIR)/$(PROG)$(BIN_SUFFIX) ./

.PHONY: dep
dep:
	dep ensure

.PHONY: test
test:
	go test -race ./...

.PHONY: check
check:
	gofmt -l -s -d pkg/
	go tool vet pkg/

.PHONY: clean
clean:
	$(RM) $(OUT_DIR)/$(PROG)$(BIN_SUFFIX)

.PHONY: uninstall
uninstall:
	$(RM) $(GOPATH)/bin/$(PROG)$(BIN_SUFFIX)

.PHONY: install
install:
	go install
