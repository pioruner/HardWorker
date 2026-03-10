# HardWorker

Debuger - go install github.com/go-delve/delve/cmd/dlv@latest

билд
 go build -ldflags "-s -w -H=windowsgui -extldflags=-static" .

## Wails Apps

Теперь проект разделен на два отдельных desktop-приложения:

- `apps/akip` - рабочее приложение для AKIP (основной UI, логи, графики)
- `apps/visco` - отдельное приложение для вискозиметра (тренды, таблица, CSV, логи)

Команды:

- `make akip-dev` / `make akip-build`
- `make visco-dev` / `make visco-build`
