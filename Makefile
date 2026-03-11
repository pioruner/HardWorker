# Makefile (Windows build with icon)

GO=go
APP=HardWorker
ICON=assets/icon.ico

AKIP_APP_DIR=apps/akip
VISCO_APP_DIR=apps/visco
UPLOADER_CONFIG?=release/windows-toolchain-prefilled-cloudru-20260311/uploader/uploader.local.json
VERSION?=dev
VERSION_DIR?=release/versions

define resolve_version
	@set -e; \
	mkdir -p $(VERSION_DIR); \
	version_file="$(VERSION_DIR)/$(1)-$(2)-$(3).version"; \
	version="$(VERSION)"; \
	if [ "$$version" = "dev" ] || [ -z "$$version" ]; then \
		today="$$(date +%Y.%m.%d)"; \
		if [ -f "$$version_file" ]; then \
			last="$$(cat "$$version_file")"; \
		else \
			last=""; \
		fi; \
		case "$$last" in \
			$$today.*) \
				rev="$${last##*.}"; \
				case "$$rev" in ''|*[!0-9]*) rev=0 ;; esac; \
				version="$$today.$$((rev + 1))" ;; \
			*) \
				seed="$$(date +%H%M%S | sed 's/^0*//')"; \
				if [ -z "$$seed" ]; then seed=1; fi; \
				version="$$today.$$seed" ;; \
		esac; \
	fi; \
	echo "$$version" > .release-version.tmp
endef

define publish_release
	@set -e; \
	version="$$(cat .release-version.tmp)"; \
	version_file="$(VERSION_DIR)/$(1)-$(2)-$(3).version"; \
	echo "Publishing $(1) $$version for $(2)/$(3)"; \
	$(GO) run ./cmd/release-publish -config $(UPLOADER_CONFIG) -app $(1) -version "$$version" -platform $(2) -arch $(3) -source $(4) -notes "$(5) $$version"; \
	printf '%s\n' "$$version" > "$$version_file"; \
	rm -f .release-version.tmp
endef

define cleanup_resolved_version
	@rm -f .release-version.tmp
endef

define require_resolved_version
	@if [ ! -f .release-version.tmp ]; then \
		echo "internal error: release version was not resolved"; \
		exit 1; \
	fi
endef

.PHONY: run build clean proto akip-dev akip-build akip-build-windows akip-release-windows visco-dev visco-build visco-build-windows visco-release-windows windows-build windows-release

run:
	$(GO) run .

build:
	@echo Creating resource file...
	echo id ICON "$(ICON)" > $(APP).rc
	echo GLFW_ICON ICON "$(ICON)" >> $(APP).rc

	@echo Compiling resource...
	windres $(APP).rc -O coff -o $(APP).syso

	@echo Building executable...
	$(GO) build -ldflags "-s -w -H=windowsgui" -o $(APP).exe .

	@echo Cleaning temp files...
	rm -f $(APP).rc
	rm -f $(APP).syso

clean:
	rm -f $(APP).exe

proto:
	protoc --go_out=./pkg/proto --go_opt=paths=source_relative --go-grpc_out=./pkg/proto --go-grpc_opt=paths=source_relative grpc.proto

akip-dev:
	cd $(AKIP_APP_DIR) && $$(go env GOPATH)/bin/wails dev

akip-build:
	cd $(AKIP_APP_DIR) && $$(go env GOPATH)/bin/wails build -clean

akip-build-windows:
	cd $(AKIP_APP_DIR) && $$(go env GOPATH)/bin/wails build -platform windows/amd64 -clean

akip-release-windows: akip-build-windows
	$(call resolve_version,akip,windows,amd64)
	$(call require_resolved_version)
	$(call publish_release,akip,windows,amd64,./$(AKIP_APP_DIR)/build/bin/akip-wails-prototype.exe,AKIP release)

visco-dev:
	cd $(VISCO_APP_DIR) && $$(go env GOPATH)/bin/wails dev

visco-build:
	cd $(VISCO_APP_DIR) && $$(go env GOPATH)/bin/wails build -clean

visco-build-windows:
	cd $(VISCO_APP_DIR) && $$(go env GOPATH)/bin/wails build -platform windows/amd64 -clean

visco-release-windows: visco-build-windows
	$(call resolve_version,visco,windows,amd64)
	$(call require_resolved_version)
	$(call publish_release,visco,windows,amd64,./$(VISCO_APP_DIR)/build/bin/hardworker-visco.exe,VISCO release)

windows-build: akip-build-windows visco-build-windows

windows-release: akip-release-windows visco-release-windows
