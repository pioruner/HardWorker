# release-publish

Console Go uploader for desktop app releases.

What it does:

- accepts a build directory or a ready `.zip`;
- creates a zip if needed;
- calculates `sha256`;
- uploads the artifact to object storage;
- updates `manifest.json` for the target app/platform/arch.

It is not tied to HardWorker apps. Any project can use it as long as it follows the same manifest format and provides a JSON config.

## Run

```bash
go run ./cmd/release-publish \
  -config ./config/uploader.local.json \
  -app myapp \
  -version 1.2.3 \
  -platform windows \
  -arch amd64 \
  -source ./release/myapp
```

Windows example with built `.exe`:

```powershell
.\release-publish.exe `
  -config .\uploader.local.json `
  -app akip `
  -version 2026.03.11.2 `
  -platform windows `
  -arch amd64 `
  -source C:\Builds\AKIP\akip-wails-prototype.exe `
  -notes "AKIP release 2026.03.11.2"
```

The command does the following:

1. reads `uploader.local.json`
2. determines where to upload artifact and manifest
3. zips the source automatically if the source is not already a `.zip`
4. calculates `sha256`
5. uploads the artifact
6. updates `manifest.json`

Successful output includes:

- published app id and version
- artifact URL
- `sha256`
- manifest update time

## Required flags

- `-app`: logical application id in manifest
- `-version`: release version
- `-source`: folder with build output or ready zip file

## Optional flags

- `-config`: config path, otherwise default config lookup is used
- `-platform`: default `windows`
- `-arch`: default `amd64`
- `-file-name`: override artifact file name
- `-notes`: release notes string stored in manifest

## Config

Use [config/uploader.example.json](/Users/cim/Documents/Projects/HardWorker/config/uploader.example.json) as a template.

The uploader config must contain:

- storage credentials
- bucket location
- manifest settings
- app list with valid app IDs

For reuse in another repository the minimal path is:

1. copy `cmd/release-publish`
2. copy `pkg/updater`
3. create a local config JSON with storage credentials
