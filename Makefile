# Makefile

# Указываем переменные
GO=go
MAIN=main.go

# Цели
.PHONY: run build

# Команда для запуска приложения
run:
	$(GO) run $(MAIN)

# Команда для сборки приложения
build:
	$(GO) build -ldflags "-s -w -H=windowsgui -extldflags=-static" .
