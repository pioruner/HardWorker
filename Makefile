# Makefile (Windows build with icon)

GO=go
APP=HardWorker
ICON=assets/icon.ico

.PHONY: run build clean

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
	del $(APP).rc
	del $(APP).syso

clean:
	del $(APP).exe