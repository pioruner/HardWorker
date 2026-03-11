# Makefile (Windows build with icon)

GO=go
APP=HardWorker
ICON=assets/icon.ico

AKIP_APP_DIR=apps/akip
VISCO_APP_DIR=apps/visco
UPLOADER_CONFIG?=release/windows-toolchain-prefilled-cloudru-20260311/uploader/uploader.local.json
VERSION?=dev

define require_version
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "Set VERSION, for example: make $@ VERSION=2026.03.11.2"; \
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
	$(call require_version)
	$(GO) run ./cmd/release-publish -config $(UPLOADER_CONFIG) -app akip -version $(VERSION) -platform windows -arch amd64 -source ./$(AKIP_APP_DIR)/build/bin/akip-wails-prototype.exe

visco-dev:
	cd $(VISCO_APP_DIR) && $$(go env GOPATH)/bin/wails dev

visco-build:
	cd $(VISCO_APP_DIR) && $$(go env GOPATH)/bin/wails build -clean

visco-build-windows:
	cd $(VISCO_APP_DIR) && $$(go env GOPATH)/bin/wails build -platform windows/amd64 -clean

visco-release-windows: visco-build-windows
	$(call require_version)
	$(GO) run ./cmd/release-publish -config $(UPLOADER_CONFIG) -app visco -version $(VERSION) -platform windows -arch amd64 -source ./$(VISCO_APP_DIR)/build/bin/hardworker-visco.exe

windows-build: akip-build-windows visco-build-windows

windows-release: akip-release-windows visco-release-windows
