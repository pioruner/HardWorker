# HardWorker Full File Map

## Scope
Этот файл описывает текущую структуру репозитория после разделения Wails-направления на отдельные прикладные и сервисные desktop-приложения:
- `apps/akip` - приложение для AKIP
- `apps/visco` - приложение для вискозиметра
- `apps/configmaster` - UI для генерации конфигов toolchain
- `apps/updater` - клиент обновления для конечной машины
- `cmd/release-publish` - консольный uploader релизов

Legacy-часть (`main.go`, `pkg/*`, `lv/*`) сохранена отдельно и не смешивается с Wails-приложениями.

## Legend
- `legacy`: старое giu-приложение в корне репозитория
- `wails-akip`: новое Wails-приложение AKIP
- `wails-visco`: новое Wails-приложение VISCO
- `shared`: общие ресурсы/документы
- `generated`: автогенерация/артефакты

## A. Root-level files
- `Makefile`: команды legacy + команды сборки/релиза для приложений (`shared`, build)
- `README.md`: верхнеуровневое описание репозитория (`shared`, docs)
- `KODA.md`: инженерные заметки по проекту (`shared`, docs)
- `go.mod`, `go.sum`, `main.go`: legacy точка входа и зависимости (`legacy`)
- `grpc.proto`: protobuf-контракт legacy AKIP ветки (`legacy`, contract)

## B. Shared assets/docs
- `assets/*`: иконки/шрифты legacy (`legacy`, ui-asset)
- `docs/wails-akip-blueprint.md`: правила для Wails-приложений (`shared`, architecture)
- `docs/wails-akip-project-structure.md`: структура Wails AKIP приложения (`wails-akip`, docs)
- `docs/project-full-file-map.md`: этот файл (`shared`, docs)
- `docs/updater-release-flow.md`: схема `ConfigMaster -> Uploader -> Updater` (`shared`, docs)
- `config/*.example.json`: примеры конфигов toolchain (`shared`, config)
- `pkg/updater/*`: общий код release/update flow (`shared`, updater-core)

## C. Legacy application (`pkg/*`, `lv/*`)

### C1. Core
- `pkg/app/main.go`: runtime-контекст/события (`legacy`)
- `pkg/ui/main.go`: оболочка вкладок/GUI (`legacy`)
- `pkg/tray/main.go`: системный трей (`legacy`)
- `pkg/logger/logger.go`: логирование (`legacy`)

### C2. Legacy modules
- `pkg/akip/*`: старый AKIP модуль (TCP, unpack, UI)
- `pkg/visko/*`: старый VISKO модуль (Modbus, UI)
- `pkg/setts/*`: настройки
- `pkg/loger/*`: legacy вкладка логов
- `pkg/proto/*`: legacy protobuf/gRPC runtime

### C3. LabVIEW
- `lv/*`: проект и RPC-сообщения для LabVIEW интеграции

## D. Wails app: AKIP (`apps/akip/*`)

### D1. Backend/entry
- `apps/akip/main.go`: Wails app options/lifecycle (`wails-akip`)
- `apps/akip/app.go`: bindings для frontend (`wails-akip`)
- `apps/akip/akip_service.go`: AKIP сервис (TCP polling, parser, calc, autosearch, CSV, persistence) (`wails-akip`)
- `apps/akip/grpc.go`: gRPC endpoint для текущих данных (`wails-akip`)
- `apps/akip/logs.go`: ring buffer логов (`wails-akip`)
- `apps/akip/go.mod`, `apps/akip/go.sum`: зависимости AKIP приложения
- `apps/akip/wails.json`: Wails CLI конфиг

### D2. Frontend
- `apps/akip/frontend/src/App.tsx`: UI AKIP + вкладка логов
- `apps/akip/frontend/src/store/akipStore.ts`: Zustand store snapshot/controls
- `apps/akip/frontend/src/App.scss`, `style.css`: стили
- `apps/akip/frontend/src/main.tsx`: bootstrap

### D3. Generated/build
- `apps/akip/frontend/wailsjs/*`: Wails JS bindings (`generated`)
- `apps/akip/build/*`: build templates/resources
- `apps/akip/build/bin/*`: выходные бинарники (`generated`)

## E. Wails app: VISCO (`apps/visco/*`)

### E1. Backend/entry
- `apps/visco/main.go`: Wails app options/lifecycle (`wails-visco`)
- `apps/visco/app.go`: bindings VISCO (`GetSnapshot`, `ApplyControls`, `GetLogs`, `SetCursorIndex`, `ClearRows`, `ExportRows`) (`wails-visco`)
- `apps/visco/visko_service.go`: VISCO сервис (Modbus polling, rows, cursor, CSV export, persistence) (`wails-visco`)
- `apps/visco/logs.go`: ring buffer логов (`wails-visco`)
- `apps/visco/go.mod`, `apps/visco/go.sum`: зависимости VISCO приложения
- `apps/visco/wails.json`: Wails CLI конфиг

### E2. Frontend
- `apps/visco/frontend/src/App.tsx`: VISCO UI (тренды, таблица, курсор, CSV) + вкладка логов
- `apps/visco/frontend/src/App.scss`, `style.css`: стили
- `apps/visco/frontend/src/main.tsx`: bootstrap

### E3. Generated/build
- `apps/visco/frontend/wailsjs/*`: Wails JS bindings (`generated`)
- `apps/visco/build/*`: build templates/resources
- `apps/visco/build/bin/*`: выходные бинарники (`generated`)

## F. Historical note
- Исторический Wails прототип удалён после миграции.
- Текущая разработка ведётся только в `apps/*`.

## G. Update toolchain

### G1. ConfigMaster (`apps/configmaster/*`)
- `apps/configmaster/main.go`: Wails entrypoint (`toolchain-configmaster`)
- `apps/configmaster/app.go`: bindings + выбор папок (`toolchain-configmaster`)
- `apps/configmaster/service.go`: хранение master config + генерация `uploader.local.json` и `updater.local.json`
- `apps/configmaster/frontend/src/App.tsx`: UI настройки toolchain
- `apps/configmaster/frontend/src/App.scss`: стили

### G2. Updater (`apps/updater/*`)
- `apps/updater/main.go`: Wails entrypoint (`toolchain-updater`)
- `apps/updater/app.go`: bindings обновления/запуска (`toolchain-updater`)
- `apps/updater/service.go`: чтение `updater.local.json`, проверка версии, установка, запуск
- `apps/updater/frontend/src/App.tsx`: клиентский UI без инфраструктурных настроек
- `apps/updater/frontend/src/App.scss`: стили

### G3. Uploader (`cmd/release-publish/*`)
- `cmd/release-publish/main.go`: CLI публикации релиза
- `cmd/release-publish/README.md`: инструкция по использованию CLI

## H. Build commands
- `make akip-dev`, `make akip-build`
- `make visco-dev`, `make visco-build`
- `make akip-release-windows VERSION=...`
- `make visco-release-windows VERSION=...`
- `make windows-release VERSION=...`

## I. Impact map
- Изменения AKIP UI/контракта: `apps/akip/app.go`, `apps/akip/akip_service.go`, `apps/akip/frontend/src/*`
- Изменения VISCO UI/контракта: `apps/visco/app.go`, `apps/visco/visko_service.go`, `apps/visco/frontend/src/*`
- Лог-модуль обязателен в обоих приложениях: `apps/*/logs.go` + вкладка `Логи` на frontend
- Изменения update toolchain: `apps/configmaster/*`, `apps/updater/*`, `cmd/release-publish/*`, `pkg/updater/*`, `config/*.example.json`

## J. Safe-to-regenerate
- `apps/*/frontend/wailsjs/**`
- `apps/*/build/**` (шаблоны/артефакты)
- `apps/*/frontend/dist/**`
- `apps/*/frontend/node_modules/**`
