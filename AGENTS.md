# AGENTS.md

## Purpose

This file is a bootstrap instruction layer for new threads.
It is guidance to lean on, not a rigid law.
It must not become a second source of project truth.

## Canonical Project Context

Read first:

1. `docs/project-context.md`

Use code as the next source of truth after that.

## Working Assumptions

- Default assumption: active product work targets `apps/*`, not legacy root code.
- Prefer targeted reads over repo-wide scans.
- If the user references HardWorker generally, start with `docs/project-context.md`.
- If the task changes reusable project knowledge, normally update `docs/project-context.md` in the same task.

## Fast File Entry Points

- AKIP:
  - `apps/akip/app.go`
  - `apps/akip/akip_service.go`
  - `apps/akip/frontend/src/App.tsx`
  - `apps/akip/frontend/src/store/akipStore.ts`
- VISCO:
  - `apps/visco/app.go`
  - `apps/visco/visko_service.go`
  - `apps/visco/frontend/src/App.tsx`
- Release/update flow:
  - `apps/configmaster/*`
  - `apps/updater/*`
  - `cmd/release-publish/*`

## Documentation Rule

- Do not create parallel project docs in `docs/` unless the user explicitly asks.
- Keep unique project knowledge in `docs/project-context.md`.
- `README.md` may link to the canonical doc, but should not hold competing detailed facts.
- Before considering a task complete, explicitly check whether `docs/project-context.md` should be updated.
- If the task changed reusable knowledge, architecture, workflows, release/update behavior, bugfix methodology, or app map, the default action is to update `docs/project-context.md` in the same task before closing it.
