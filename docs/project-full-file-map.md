# HardWorker Full File Map

## Scope
Этот файл описывает **все tracked-файлы репозитория** (по `git ls-files`) и их роль.
Цель: безопасно мигрировать с legacy-структуры на новую Wails-архитектуру, понимая зависимости и зоны риска.

## Legend
- `legacy`: используется старым приложением в корне (`main.go` + `pkg/*` + `lv/*`).
- `prototype`: используется новым Wails-проектом в `experiments/akip-wails-prototype`.
- `shared`: может использоваться обеими ветками развития.
- `generated`: сгенерированный/служебный файл.

## A. Root-level files
- `.codex/environments/environment.toml`: локальная конфигурация окружения Codex (`shared`, infra).
- `.gitignore`: глобальные игноры репозитория (`shared`, infra).
- `.vscode/launch.json`: локальные конфиги запуска/отладки VSCode (`shared`, infra).
- `KODA.md`: проектная документация/заметки (`shared`, docs).
- `Makefile`: команды сборки/утилиты legacy-проекта (`legacy`, build).
- `README.md`: общий обзор репозитория (`shared`, docs).
- `go.mod`: Go-модуль legacy-приложения (`legacy`, build).
- `go.sum`: lockfile зависимостей legacy (`legacy`, build).
- `grpc.proto`: protobuf-контракт сервиса Akip (`shared`, contract).
- `main.go`: точка входа legacy desktop-приложения (`legacy`, runtime).

## B. Root assets
- `assets/icon.ico`: иконка Windows legacy-приложения (`legacy`, ui-asset).
- `assets/icon.png`: PNG иконка/ресурс legacy (`legacy`, ui-asset).
- `assets/inter.ttf`: шрифт для UI (`legacy`, ui-asset).

## C. Root docs
- `docs/wails-akip-blueprint.md`: blueprint нового подхода (стек, архитектура, правила) (`prototype`, docs).
- `docs/wails-akip-project-structure.md`: структура только Wails-прототипа (`prototype`, docs).
- `docs/project-full-file-map.md`: этот полный файл-реестр (`shared`, docs).

## D. Legacy Go application (`pkg/*`)

### D1. Core/runtime
- `pkg/app/main.go`: глобальный runtime-контекст, синхронизация модулей (`legacy`, core).
- `pkg/ui/main.go`: главный UI-контур legacy-приложения, вкладки/маршрутизация модулей (`legacy`, core-ui).
- `pkg/tray/main.go`: системный трей/интеграция с оболочкой (`legacy`, platform).
- `pkg/logger/logger.go`: общий логгер backend/модулей (`legacy`, infra).

### D2. AKIP module
- `pkg/akip/main.go`: модель/состояние AKIP-модуля, load/save, registration-loop (`legacy`, domain).
- `pkg/akip/commands.go`: связь с прибором, polling, unpack wave, calc/autosearch (`legacy`, domain).
- `pkg/akip/ui.go`: legacy UI AKIP-модуля на `giu` (`legacy`, ui).
- `pkg/akip/grpc.go`: gRPC слой legacy AKIP (исторически с недоработкой `Data`) (`legacy`, integration).

### D3. VISKO module
- `pkg/visko/main.go`: состояние/инициализация Visko-модуля (`legacy`, domain).
- `pkg/visko/commands.go`: команды/обмен для Visko (`legacy`, domain).
- `pkg/visko/ui.go`: UI Visko на `giu` (`legacy`, ui).

### D4. Settings module
- `pkg/setts/main.go`: состояние/логика настроек (`legacy`, domain).
- `pkg/setts/ui.go`: UI настроек (`legacy`, ui).

### D5. System log UI module
- `pkg/loger/main.go`: подключение legacy-вкладки системного лога (`legacy`, ui).
- `pkg/loger/logui.go`: буфер log-строк, уровень/цвет, writer для UI (`legacy`, infra-ui).
- `pkg/loger/ui.go`: отрисовка цветного списка логов в legacy UI (`legacy`, ui).

### D6. Legacy proto runtime
- `pkg/proto/main.go`: standalone запуск gRPC в legacy окружении (`legacy`, integration).
- `pkg/proto/grpc.pb.go`: protobuf types legacy (`generated`, contract).
- `pkg/proto/grpc_grpc.pb.go`: gRPC service interfaces legacy (`generated`, contract).

## E. LabVIEW integration (`lv/*`)
- `lv/Proto.aliases`: aliases/пути LabVIEW-проекта (`legacy`, labview-config).
- `lv/Proto.lvlps`: LabVIEW project support file (`legacy`, labview-config).
- `lv/Proto.lvproj`: основной LabVIEW проект (`legacy`, labview-project).
- `lv/Test.vi`: тестовый VI (`legacy`, labview-test).

### E1. LabVIEW Proto_client library
- `lv/Proto_client/Proto_client.lvlib`: библиотека клиента protobuf/gRPC в LabVIEW (`legacy`, labview-lib).
- `lv/Proto_client/Client API/Create Client.vi`: создание клиента (`legacy`, labview-api).
- `lv/Proto_client/Client API/Destroy Client.vi`: освобождение клиента (`legacy`, labview-api).
- `lv/Proto_client/RPC Service/Akip/Akip Data.vi`: вызов RPC `Akip.Data` (`legacy`, labview-rpc).

### E2. LabVIEW RPC Messages
- `lv/Proto_client/RPC Messages/Register gRPC Messages.vi`: регистрация message-карт (`legacy`, labview-rpc).
- `lv/Proto_client/RPC Messages/grpcservice_AkipRequest.ctl`: rich request control (`legacy`, labview-rpc).
- `lv/Proto_client/RPC Messages/grpcservice_AkipRequest_Flat.ctl`: flat request control (`legacy`, labview-rpc).
- `lv/Proto_client/RPC Messages/grpcservice_AkipResponse.ctl`: rich response control (`legacy`, labview-rpc).
- `lv/Proto_client/RPC Messages/grpcservice_AkipResponse_Flat.ctl`: flat response control (`legacy`, labview-rpc).
- `lv/Proto_client/RPC Messages/RichToFlatgrpcservice_AkipRequest.vi`: сериализация rich->flat request (`legacy`, labview-rpc).
- `lv/Proto_client/RPC Messages/FlatToRichgrpcservice_AkipRequest.vi`: десериализация flat->rich request (`legacy`, labview-rpc).
- `lv/Proto_client/RPC Messages/RichToFlatgrpcservice_AkipResponse.vi`: сериализация rich->flat response (`legacy`, labview-rpc).
- `lv/Proto_client/RPC Messages/FlatToRichgrpcservice_AkipResponse.vi`: десериализация flat->rich response (`legacy`, labview-rpc).

## F. Wails prototype (`experiments/akip-wails-prototype/*`)

### F1. Go backend and app entry
- `experiments/akip-wails-prototype/main.go`: Wails app options, window behavior, lifecycle hooks (`prototype`, runtime).
- `experiments/akip-wails-prototype/app.go`: Wails bindings (`GetSnapshot`, `ApplyControls`, `GetLogs`, etc.) (`prototype`, bridge).
- `experiments/akip-wails-prototype/akip_service.go`: основной domain-service AKIP (TCP, poll, calc, CSV, state) (`prototype`, domain).
- `experiments/akip-wails-prototype/grpc.go`: gRPC server prototype (`prototype`, integration).
- `experiments/akip-wails-prototype/logs.go`: in-memory log buffer + log model (`prototype`, infra).
- `experiments/akip-wails-prototype/go.mod`: Go module prototype (`prototype`, build).
- `experiments/akip-wails-prototype/go.sum`: Go lockfile prototype (`prototype`, build).
- `experiments/akip-wails-prototype/wails.json`: Wails CLI config (`prototype`, build).
- `experiments/akip-wails-prototype/README.md`: базовый README шаблона Wails (`prototype`, docs).
- `experiments/akip-wails-prototype/.gitignore`: игноры внутри prototype (`prototype`, infra).
- `experiments/akip-wails-prototype/akip-wails-prototype`: локальный бинарник/артефакт (mac test) (`generated`, artifact).

### F2. Prototype protobuf
- `experiments/akip-wails-prototype/proto/grpc.pb.go`: protobuf types prototype (`generated`, contract).
- `experiments/akip-wails-prototype/proto/grpc_grpc.pb.go`: gRPC service interfaces prototype (`generated`, contract).

### F3. Frontend app
- `experiments/akip-wails-prototype/frontend/index.html`: Vite HTML entry (`prototype`, frontend).
- `experiments/akip-wails-prototype/frontend/package.json`: npm scripts + dependencies (`prototype`, frontend-build).
- `experiments/akip-wails-prototype/frontend/package-lock.json`: npm lockfile (`generated`, frontend-build).
- `experiments/akip-wails-prototype/frontend/package.json.md5`: контрольная сумма зависимостей для Wails tooling (`generated`, frontend-build).
- `experiments/akip-wails-prototype/frontend/tsconfig.json`: TS config frontend (`prototype`, frontend-build).
- `experiments/akip-wails-prototype/frontend/tsconfig.node.json`: TS config для Vite/node-side (`prototype`, frontend-build).
- `experiments/akip-wails-prototype/frontend/vite.config.ts`: Vite config (`prototype`, frontend-build).
- `experiments/akip-wails-prototype/frontend/src/main.tsx`: bootstrap React root (`prototype`, frontend).
- `experiments/akip-wails-prototype/frontend/src/App.tsx`: основной UI, вкладки AKIP/Логи, команды пользователю (`prototype`, frontend-ui).
- `experiments/akip-wails-prototype/frontend/src/App.scss`: дизайн-система и layout (`prototype`, frontend-ui).
- `experiments/akip-wails-prototype/frontend/src/style.css`: глобальные базовые стили (`prototype`, frontend-ui).
- `experiments/akip-wails-prototype/frontend/src/vite-env.d.ts`: типы окружения Vite (`generated`, frontend).
- `experiments/akip-wails-prototype/frontend/src/store/akipStore.ts`: Zustand store/models (`prototype`, frontend-state).

### F4. Frontend assets
- `experiments/akip-wails-prototype/frontend/src/assets/fonts/OFL.txt`: лицензия шрифта (`prototype`, asset-license).
- `experiments/akip-wails-prototype/frontend/src/assets/fonts/nunito-v16-latin-regular.woff2`: шрифт UI (`prototype`, ui-asset).
- `experiments/akip-wails-prototype/frontend/src/assets/images/logo-universal.png`: базовый шаблонный логотип (`prototype`, ui-asset).

### F5. Wails JS bindings/runtime
- `experiments/akip-wails-prototype/frontend/wailsjs/go/main/App.d.ts`: TS-описание backend методов (`generated`, bridge).
- `experiments/akip-wails-prototype/frontend/wailsjs/go/main/App.js`: JS bridge вызовов backend (`generated`, bridge).
- `experiments/akip-wails-prototype/frontend/wailsjs/go/models.ts`: TS-модели Go-структур (`generated`, bridge).
- `experiments/akip-wails-prototype/frontend/wailsjs/runtime/package.json`: runtime package meta (`generated`, bridge).
- `experiments/akip-wails-prototype/frontend/wailsjs/runtime/runtime.d.ts`: TS typings runtime API (`generated`, bridge).
- `experiments/akip-wails-prototype/frontend/wailsjs/runtime/runtime.js`: runtime API Wails (`generated`, bridge).

### F6. Wails build templates/resources
- `experiments/akip-wails-prototype/build/README.md`: справка по build-ресурсам (`prototype`, build-doc).
- `experiments/akip-wails-prototype/build/appicon.png`: app icon source (`prototype`, build-asset).
- `experiments/akip-wails-prototype/build/darwin/Info.plist`: plist production template macOS (`prototype`, build-template).
- `experiments/akip-wails-prototype/build/darwin/Info.dev.plist`: plist dev template macOS (`prototype`, build-template).
- `experiments/akip-wails-prototype/build/windows/icon.ico`: windows icon template (`prototype`, build-asset).
- `experiments/akip-wails-prototype/build/windows/info.json`: windows metadata template (`prototype`, build-template).
- `experiments/akip-wails-prototype/build/windows/wails.exe.manifest`: windows manifest template (`prototype`, build-template).
- `experiments/akip-wails-prototype/build/windows/installer/project.nsi`: NSIS installer template (`prototype`, build-template).
- `experiments/akip-wails-prototype/build/windows/installer/wails_tools.nsh`: NSIS helper macros (`prototype`, build-template).

## G. Dependency/impact map for migration
- If changing `grpc.proto`:
  - regenerate both `pkg/proto/*` and `experiments/.../proto/*` if обе ветки должны быть совместимы.
- If changing AKIP business logic:
  - legacy: `pkg/akip/*`;
  - prototype: `experiments/.../akip_service.go`;
  - verify UI fields and gRPC `Data()` outputs.
- If changing log UX:
  - prototype UI: `frontend/src/App.tsx`, `frontend/src/App.scss`;
  - backend events: `akip_service.go`, `grpc.go`, `logs.go`.
- If changing snapshot contract:
  - Go structs in `akip_service.go`;
  - TS models/store in `frontend/src/store/akipStore.ts`;
  - validate regenerated `frontend/wailsjs/go/models.ts`.

## H. Files usually safe to regenerate (do not treat as source of truth)
- `experiments/akip-wails-prototype/frontend/wailsjs/**`
- `experiments/akip-wails-prototype/build/**` templates (if not custom-edited)
- compiled artifacts: `build/bin/*`, `frontend/dist/*`, `frontend/node_modules/*`

