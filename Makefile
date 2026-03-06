# Makefile (Windows build with icon)

GO=go
APP=HardWorker
ICON=assets/icon.ico

.PHONY: run build clean proto

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