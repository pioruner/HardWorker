This directory stores the last successfully published release version per app/platform/arch.

Format:

- `<app>-<platform>-<arch>.version`

Current Makefile behavior:

- `make visco-release-windows` auto-generates the next version in `YYYY.MM.DD.N` format
- if a version file exists for today, `N` is incremented
- if the day changed or the file is missing, the initial suffix is seeded from current time to avoid collisions with already published releases
- after a successful publish, the version file is updated
- manual override is still possible with `VERSION=...`
