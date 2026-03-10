# HardWorker Full File Map

## Scope
Этот файл описывает текущую структуру репозитория после разделения Wails-направления на два отдельных приложения:
- `apps/akip` - приложение для AKIP
- `apps/visco` - приложение для вискозиметра

Legacy-часть (`main.go`, `pkg/*`, `lv/*`) сохранена отдельно и не смешивается с Wails-приложениями.

## Legend
- `legacy`: старое giu-приложение в корне репозитория
- `wails-akip`: новое Wails-приложение AKIP
- `wails-visco`: новое Wails-приложение VISCO
- `shared`: общие ресурсы/документы
- `generated`: автогенерация/артефакты

## A. Root-level files
- `Makefile`: команды legacy + команды для `apps/akip` и `apps/visco` (`shared`, build)
- `README.md`: верхнеуровневое описание репозитория (`shared`, docs)
- `KODA.md`: инженерные заметки по проекту (`shared`, docs)
- `go.mod`, `go.sum`, `main.go`: legacy точка входа и зависимости (`legacy`)
- `grpc.proto`: protobuf-контракт legacy AKIP ветки (`legacy`, contract)

## B. Shared assets/docs
- `assets/*`: иконки/шрифты legacy (`legacy`, ui-asset)
- `docs/wails-akip-blueprint.md`: правила для Wails-приложений (`shared`, architecture)
- `docs/wails-akip-project-structure.md`: структура Wails AKIP приложения (`wails-akip`, docs)
- `docs/project-full-file-map.md`: этот файл (`shared`, docs)

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

## G. Build commands
- `make akip-dev`, `make akip-build`
- `make visco-dev`, `make visco-build`

## H. Impact map
- Изменения AKIP UI/контракта: `apps/akip/app.go`, `apps/akip/akip_service.go`, `apps/akip/frontend/src/*`
- Изменения VISCO UI/контракта: `apps/visco/app.go`, `apps/visco/visko_service.go`, `apps/visco/frontend/src/*`
- Лог-модуль обязателен в обоих приложениях: `apps/*/logs.go` + вкладка `Логи` на frontend

## I. Safe-to-regenerate
- `apps/*/frontend/wailsjs/**`
- `apps/*/build/**` (шаблоны/артефакты)
- `apps/*/frontend/dist/**`
- `apps/*/frontend/node_modules/**`
