# Wails AKIP Project Structure

## Назначение документа
Этот документ описывает структуру `experiments/akip-wails-prototype`, роли файлов и их связи, чтобы безопасно переносить решение в основное приложение.

## 1. Корень проекта
Путь: `experiments/akip-wails-prototype`

Ключевые файлы:
- `main.go`: точка входа Wails-приложения, параметры окна, lifecycle hooks, bind backend-методов.
- `app.go`: Wails-bridge (методы, доступные frontend), проксирует вызовы в `AkipService`.
- `akip_service.go`: основная backend-логика прибора (TCP, опрос, парсинг, расчеты, state, CSV, zero reference).
- `grpc.go`: gRPC-сервер (`:50051`), отдаёт текущие вычисленные данные.
- `logs.go`: in-memory лог-буфер и модель лог-записи.
- `go.mod` / `go.sum`: зависимости Go.
- `wails.json`: конфиг Wails CLI.
- `README.md`: базовый шаблон проекта (можно дополнять/упрощать под команду).

## 2. Backend-слой (Go)
### 2.1 `akip_service.go`
Что делает:
- хранит текущее состояние `AkipService`;
- управляет TCP соединением и reconnect;
- выполняет polling данных осциллограммы;
- парсит бинарный пакет прибора;
- считает метрики (`vSpeed`, `vTime`, `volume`);
- поддерживает `ZeroVolumeReference` (дельта-измерение);
- управляет CSV-регистрацией;
- load/save JSON-состояния в `UserConfigDir()/HardWorker`.

С чем связан:
- вызывается из `app.go`;
- пишет в `logs.go` буфер;
- запускает `grpcLoop()` из `grpc.go`.

### 2.2 `grpc.go`
Что делает:
- поднимает gRPC endpoint;
- реализует `Data()` и возвращает актуальное значение объема из `AkipService`.

С чем связан:
- использует protobuf-типы из `proto/*`;
- получает данные через методы `AkipService`.

### 2.3 `logs.go`
Что делает:
- держит кольцевой буфер логов;
- нормализует формат (`time`, `level`, `message`).

С чем связан:
- backend-события пишут сюда;
- frontend читает через `GetLogs()` (через `app.go`).

### 2.4 `app.go`
Что делает:
- экспортирует методы для frontend:
  - `GetSnapshot()`
  - `ApplyControls(...)`
  - `SetRegistration(...)`
  - `ZeroVolumeReference()`
  - `GetLogs()`
- привязывает lifecycle startup/shutdown.

## 3. Protobuf/gRPC модели
Папка: `proto/`
- `grpc.pb.go`
- `grpc_grpc.pb.go`

Это сгенерированные файлы protobuf/gRPC.
Рекомендуется:
- не редактировать вручную;
- при изменении `grpc.proto` перегенерировать и обновить эти файлы.

## 4. Frontend-слой
Путь: `frontend/`

Ключевые файлы:
- `src/main.tsx`: bootstrap React-приложения.
- `src/App.tsx`: основной UI (вкладки `AKIP` / `Логи`).
- `src/App.scss`: стили layout + log view + цвета уровней.
- `src/style.css`: базовые глобальные стили.
- `src/store/akipStore.ts`: Zustand store, модели snapshot/controls/logs.

С чем связан:
- импортирует bridge-методы из `frontend/wailsjs/go/main/App`.
- polling:
  - `GetSnapshot()` (данные прибора)
  - `GetLogs()` (лог-лента)

## 5. Wails bindings (автогенерация)
Путь: `frontend/wailsjs/`

Важные файлы:
- `go/main/App.d.ts`
- `go/main/App.js`
- `go/models.ts`
- `runtime/*`

Это генерируемые Wails CLI файлы.
Правило:
- можно коммитить (для стабильной типизации);
- но при изменении backend-методов они будут перегенерированы.

## 6. Сборочные/служебные директории
- `build/`: иконки, манифесты, шаблоны инсталлятора.
- `build/bin/`: выходные бинарники (`.exe`, `.app`) после `wails build`.
- `frontend/dist/`: сборка frontend.
- `frontend/node_modules/`: npm зависимости.

Для миграции в основной проект:
- `build/bin`, `frontend/dist`, `frontend/node_modules` не являются бизнес-исходниками;
- их можно пересобрать, не переносить как источник логики.

## 7. Карта зависимостей (упрощенно)
1. `main.go` -> создаёт `App` -> bind методов.
2. `app.go` -> управляет `AkipService`.
3. `AkipService`:
   - TCP/парсинг/расчеты/CSV/state;
   - пишет в `LogBuffer`;
   - поднимает gRPC.
4. Frontend:
   - читает snapshot/logs;
   - отправляет controls/команды.

## 8. Что переносить в основное приложение в первую очередь
1. `akip_service.go`, `grpc.go`, `logs.go`, `app.go` (ядро backend).
2. `frontend/src/App.tsx`, `frontend/src/store/akipStore.ts`, `frontend/src/App.scss` (UI + state).
3. `proto/*` (если сохраняется текущий gRPC контракт).
4. `main.go` параметры окна и bind/lifecycle.

## 9. Что чаще всего меняется при переходе в production
- название модуля в `go.mod`;
- адреса/порты по умолчанию;
- поля snapshot/controls под реальное железо;
- логика автопоиска и расчетные формулы;
- UX панели логов (фильтрация, экспорт, лимиты).

## 10. Контрольный список перед удалением/рефакторингом файлов
Перед удалением любого файла проверь:
1. Есть ли прямой импорт/вызов из `App.tsx`, `app.go`, `main.go`.
2. Не участвует ли файл в автогенерации bindings (`frontend/wailsjs`).
3. Не используется ли тип в `proto/*` и gRPC-сервисе.
4. Не сломает ли удаление сохранение state/CSV/логирование.

