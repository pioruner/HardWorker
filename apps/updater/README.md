# HardWorker Updater

Desktop updater for HardWorker apps. It reads local JSON config, fetches `manifest.json`, downloads the latest zip artifact, verifies `sha256`, and swaps the install directory.

Config lookup order:

1. `HARDWORKER_UPDATER_CONFIG`
2. `./updater.local.json` near the executable
3. `./config/updater.local.json`
4. workspace `config/updater.local.json`

Publisher CLI:

```bash
go run ./cmd/release-publish -config ./config/updater.local.json -app akip -version 0.1.0 -platform windows -arch amd64 -source ./release/akip
```

