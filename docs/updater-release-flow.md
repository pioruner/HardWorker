# Release Toolchain

## Архитектура

Система теперь делится на три отдельных приложения.

### 1. ConfigMaster

- UI-приложение для настройки проекта.
- Хранит полный мастер-конфиг локально.
- Редактирует:
  - `manifest_url`
  - S3/bucket credentials
  - список приложений
  - пути, куда генерировать конфиги
- Генерирует два файла:
  - `uploader.local.json`
  - `updater.local.json`

Папка приложения: `apps/configmaster`

### 2. Uploader

- Консольное Go-приложение без UI.
- Читает только `uploader.local.json`.
- Загружает zip-артефакт и обновляет `manifest.json`.

Папка приложения: `cmd/release-publish`

### 3. Updater

- Простое UI-приложение для конечного пользователя.
- Читает только `updater.local.json`.
- Не содержит ключей доступа в UI, но может читать их из файла конфигурации для приватного bucket.
- Показывает:
  - текущую версию
  - версию на сервере
  - статус связи
  - кнопки `Проверить`, `Обновить`, `Запустить`

Папка приложения: `apps/updater`

## Поток работы

1. Запустить `ConfigMaster`.
2. Настроить bucket, manifest URL и список приложений.
3. Нажать `Сохранить мастер`.
4. Нажать `Сгенерировать файлы`.
5. Положить `uploader.local.json` рядом с `Uploader`.
6. Положить `updater.local.json` рядом с `Updater`.
7. Собрать нужное приложение.
8. Опубликовать его через `Uploader`.
9. Передать заказчику `Updater` вместе с `updater.local.json`.

## Публикация через Uploader

Базовая команда:

```bash
release-publish.exe ^
  -config .\uploader.local.json ^
  -app akip ^
  -version 2026.03.11.2 ^
  -platform windows ^
  -arch amd64 ^
  -source C:\Builds\AKIP\akip-wails-prototype.exe ^
  -notes "AKIP release 2026.03.11.2"
```

Что означает каждый флаг:

- `-config`: путь к `uploader.local.json`
- `-app`: ID приложения из конфига, например `akip` или `visco`
- `-version`: версия релиза, которая попадет в `manifest.json`
- `-platform`: целевая платформа
- `-arch`: целевая архитектура
- `-source`: путь к `.exe`, `.zip` или папке со сборкой
- `-notes`: заметка о релизе

Пример для VISCO:

```bash
release-publish.exe ^
  -config .\uploader.local.json ^
  -app visco ^
  -version 2026.03.11.2 ^
  -platform windows ^
  -arch amd64 ^
  -source C:\Builds\VISCO\hardworker-visco.exe ^
  -notes "VISCO release 2026.03.11.2"
```

После успешной публикации uploader выводит:

- ID и версию опубликованного приложения
- итоговый URL артефакта
- `sha256`
- время обновления `manifest.json`

## Автоверсия через Makefile

Для `HardWorker` ручной `VERSION=...` теперь не обязателен.

Команда:

```bash
make visco-release-windows
```

делает следующее:

1. собирает `VISCO` под `windows/amd64`
2. определяет следующую версию в формате `YYYY.MM.DD.N`
3. публикует артефакт в облако через `release-publish`
4. обновляет локальный файл версии только после успешной публикации

Файлы версий лежат в `release/versions/`, например:

- `release/versions/visco-windows-amd64.version`
- `release/versions/akip-windows-amd64.version`

Правило инкремента:

- если последняя версия была выпущена сегодня, увеличивается только хвост `N`
- если сегодня новый день или файла ещё нет, начальный хвост берётся из текущего времени, чтобы не столкнуться с уже опубликованными сегодня релизами

При необходимости версию всё ещё можно задать вручную:

```bash
make visco-release-windows VERSION=2026.03.11.7
```

## Почему так

- конечный пользователь не видит инфраструктурных настроек;
- секреты остаются в `ConfigMaster` и `Uploader`, а не в `Updater`;
- при приватном bucket секреты могут также храниться в `updater.local.json`, но не показываются конечному пользователю в интерфейсе;
- один мастер-конфиг может описывать несколько приложений;
- `Uploader` и `Updater` становятся простыми и предсказуемыми.

## Файлы примеров

- [config/configmaster.example.json](/Users/cim/Documents/Projects/HardWorker/config/configmaster.example.json)
- [config/uploader.example.json](/Users/cim/Documents/Projects/HardWorker/config/uploader.example.json)
- [config/updater.example.json](/Users/cim/Documents/Projects/HardWorker/config/updater.example.json)
