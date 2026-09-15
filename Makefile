DATE=$(shell date +%Y%m%d)
# Release builds pass VERSION= explicitly (the git tag, e.g. v1.0.0) and are
# stamped verbatim, so `--version` reports exactly the released version.
# Previously the date was appended unconditionally, turning a v1.0.0 release
# into "v1.0.0-20260729" and forcing updatechk to strip the suffix back off.
# Local/dev builds get a descriptive git-derived version with a date suffix.
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0-dev)-$(DATE)
LDFLAGS=-ldflags "-s -w -X github.com/wso2/integration-platform-tools/internal/cmd.Version=$(VERSION)"
# mcpb manifests take plain semver, without the leading "v".
MCPB_VERSION=$(patsubst v%,%,$(VERSION))
IP_OUTPUT:=bin/$(OS)-$(ARCH)/wso2-integration-platform
GEN_TRANSLATIONS_CMD:=scripts/translations/gen-translation-files.sh
PACKAGE_TRANSLATIONS_CMD:=scripts/translations/package-translations.sh

ifeq ($(OS),windows)
	IP_OUTPUT:=bin/$(OS)-$(ARCH)/wso2-integration-platform.exe
	GEN_TRANSLATIONS:=scripts\translations\gen-translation-files.bat
	PACKAGE_TRANSLATIONS:=scripts\translations\package-translations.bat
endif

COMPILE_IP_CMD:=GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) -o $(IP_OUTPUT) cmd/main.go

ifeq ($(OS),linux)
	COMPILE_IP_CMD:=CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -a -installsuffix cgo $(LDFLAGS) -o $(IP_OUTPUT) cmd/main.go
endif


MCPB_BINARY:=wso2-integration-platform
MCPB_MANIFEST:=manifest.json

ifeq ($(OS),linux)
	IP_ARCHIVE_CMD:=cd bin/$(OS)-$(ARCH) && tar -zcvf wso2-integration-platform-$(OS)-$(ARCH).tar.gz wso2-integration-platform
else ifeq ($(OS),darwin)
	IP_ARCHIVE_CMD:=cd bin/$(OS)-$(ARCH) && zip wso2-integration-platform-$(OS)-$(ARCH).zip wso2-integration-platform
	MCPB_MANIFEST:=manifest-darwin.json
else ifeq ($(OS),windows)
	IP_ARCHIVE_CMD:=cd bin/$(OS)-$(ARCH) && zip wso2-integration-platform-$(OS)-$(ARCH).zip wso2-integration-platform.exe
	MCPB_BINARY:=wso2-integration-platform.exe
	MCPB_MANIFEST:=manifest-windows.json
endif

build: package_translations
	go build $(LDFLAGS) -o bin/wso2-integration-platform cmd/main.go

run: package_translations
	go run $(LDFLAGS) cmd/main.go

test:
	go test -race -v $(shell go list ./... | grep -v integration_tests) -timeout 1200s

clean:
	rm -rf bin/*

# Cross-compile the Integration Platform branded binary
compile_ip: package_translations
	@${COMPILE_IP_CMD}

# create dist archive for Integration Platform binary
dist_ip:
	@$(IP_ARCHIVE_CMD)

# create .mcpb Claude Connector bundle
bundle_mcpb: compile_ip
	mkdir -p bin/$(OS)-$(ARCH) /tmp/mcpb-build/server
	cp $(IP_OUTPUT) /tmp/mcpb-build/server/$(MCPB_BINARY)
	# Stamp the release version into the manifest. The checked-in value is a
	# placeholder; copying it verbatim shipped every bundle as "1.0.0", and the
	# MCP host uses this field to detect connector updates.
	sed 's|"version": "[^"]*"|"version": "$(MCPB_VERSION)"|' mcpb/$(MCPB_MANIFEST) > /tmp/mcpb-build/manifest.json
	cd /tmp/mcpb-build && zip -r $(CURDIR)/bin/$(OS)-$(ARCH)/wso2-integration-platform-$(OS)-$(ARCH).mcpb manifest.json server/$(MCPB_BINARY)
	rm -rf /tmp/mcpb-build

gen_translations:
	@$(GEN_TRANSLATIONS_CMD)

package_translations:
	@$(PACKAGE_TRANSLATIONS_CMD)

run_integration_tests:
	@go test ./integration_tests/ -run TestE2EMainFlow -v -timeout 1500s $(if $(TEST),-run "$(TEST)")


extract_tools_defs:
	go run extract_tools.go
