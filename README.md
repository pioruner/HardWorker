# HardWorker

Debuger - go install github.com/go-delve/delve/cmd/dlv@latest

билд
 go build -ldflags "-s -w -H=windowsgui -extldflags=-static" .

## Wails Apps

Теперь проект содержит:

- `apps/akip` - рабочее приложение для AKIP (основной UI, логи, графики)
- `apps/visco` - отдельное приложение для вискозиметра (тренды, таблица, CSV, логи)
- `apps/rp40` - калькулятор RP40 для оценки газопроницаемости по паспортному файлу и кривой падения давления
- `apps/configmaster` - мастер-настройка update toolchain
- `apps/updater` - клиент обновления для заказчика
- `cmd/release-publish` - консольный uploader релизов

Команды:

- `make akip-dev` / `make akip-build`
- `make visco-dev` / `make visco-build`
- `make rp40-dev` / `make rp40-build`
- `make rp40-release-windows VERSION=...`
- `make akip-release-windows VERSION=...`
- `make visco-release-windows VERSION=...`
- `make windows-release VERSION=...`

## Fast Context For New Threads

- Canonical project context: [docs/project-context.md](/Users/cim/Documents/Projects/HardWorker/docs/project-context.md)
- Assistant bootstrap instructions: [AGENTS.md](/Users/cim/Documents/Projects/HardWorker/AGENTS.md)

## Update Toolchain

Текущая схема:

1. `ConfigMaster` создает и обновляет:
   - `configmaster.local.json`
   - `uploader.local.json`
   - `updater.local.json`
2. `release-publish` публикует сборку в bucket и обновляет `manifest.json`
3. `Updater` проверяет доступную версию, скачивает обновление и запускает приложение

Подробности:

- [docs/project-context.md](/Users/cim/Documents/Projects/HardWorker/docs/project-context.md)
- [cmd/release-publish/README.md](/Users/cim/Documents/Projects/HardWorker/cmd/release-publish/README.md)
