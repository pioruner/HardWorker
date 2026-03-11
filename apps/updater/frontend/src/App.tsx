import { useEffect, useMemo, useState } from "react";
import "./App.scss";
import { GetState, LaunchApp, Refresh, StartUpdate } from "../wailsjs/go/main/App";
import logoUrl from "./assets/images/logo-universal.png";

type ProgressState = {
  stage: string;
  message: string;
  percent: number;
  bytes_done: number;
  bytes_total: number;
  started_at: string;
  finished_at: string;
  download_path?: string;
};

type ManagedAppState = {
  id: string;
  label: string;
  platform: string;
  arch: string;
  install_dir: string;
  executable: string;
  current_version: string;
  latest_version: string;
  status: string;
  message: string;
  has_update: boolean;
};

type LogEntry = {
  time: string;
  level: string;
  message: string;
};

type Snapshot = {
  config_path: string;
  ready: boolean;
  busy: boolean;
  error: string;
  apps: ManagedAppState[];
  logs: LogEntry[];
  progress: ProgressState;
};

const emptyState: Snapshot = {
  config_path: "",
  ready: false,
  busy: false,
  error: "",
  apps: [],
  logs: [],
  progress: {
    stage: "idle",
    message: "Ожидание",
    percent: 0,
    bytes_done: 0,
    bytes_total: 0,
    started_at: "",
    finished_at: "",
    download_path: "",
  },
};

function App() {
  const [state, setState] = useState<Snapshot>(emptyState);
  const [activeAppID, setActiveAppID] = useState("");
  const [launchError, setLaunchError] = useState("");

  useEffect(() => {
    let active = true;
    const sync = async () => {
      try {
        const next = normalizeSnapshot(await GetState());
        if (!active) return;
        setState(next);
        if (!activeAppID && next.apps.length > 0) {
          setActiveAppID(next.apps[0].id);
        }
      } catch {
        // ignore polling hiccups
      }
    };
    void sync();
    const id = window.setInterval(sync, 900);
    return () => {
      active = false;
      window.clearInterval(id);
    };
  }, [activeAppID]);

  const selected = useMemo(
    () => state.apps.find((app) => app.id === activeAppID) ?? state.apps[0],
    [activeAppID, state.apps],
  );

  const onRefresh = async () => {
    setLaunchError("");
    setState(normalizeSnapshot(await Refresh()));
  };

  const onUpdate = async () => {
    if (!selected?.id) return;
    setLaunchError("");
    setState(normalizeSnapshot(await StartUpdate(selected.id)));
  };

  const onLaunch = async () => {
    if (!selected?.id) return;
    const error = await LaunchApp(selected.id);
    setLaunchError(error);
  };

  const logs = state.logs.slice(0, 12);
  const progressLabel =
    state.progress.bytes_total > 0
      ? `${formatBytes(state.progress.bytes_done)} / ${formatBytes(state.progress.bytes_total)}`
      : state.progress.message;

  return (
    <div className="updater-layout">
      <aside className="sidebar panel">
        <div className="brand">
          <img alt="Updater" src={logoUrl} />
          <div>
            <strong>HardWorker Updater</strong>
            <span>простое обновление и запуск</span>
          </div>
        </div>
        <div className="config-card">
          <span>Конфиг</span>
          <code>{state.config_path || "updater.local.json не найден"}</code>
        </div>
        <div className="app-list">
          {state.apps.map((app) => (
            <button
              key={app.id}
              className={`app-card ${selected?.id === app.id ? "is-active" : ""}`}
              onClick={() => setActiveAppID(app.id)}
              type="button"
            >
              <div className="app-card__head">
                <strong>{app.label || app.id}</strong>
                <span className={`status status--${app.status}`}>{statusLabel(app.status)}</span>
              </div>
              <div className="app-card__meta">
                <span>Текущая</span>
                <strong>{app.current_version || "n/a"}</strong>
              </div>
              <div className="app-card__meta">
                <span>На сервере</span>
                <strong>{app.latest_version || "n/a"}</strong>
              </div>
            </button>
          ))}
        </div>
      </aside>

      <main className="content">
        <section className="hero panel">
          <div>
            <span className="eyebrow">Update Status</span>
            <h1>{selected?.label ?? "Приложение не настроено"}</h1>
            <p>{selected?.message ?? state.error ?? "Положите рядом корректный updater.local.json, созданный в ConfigMaster."}</p>
          </div>
          <div className="hero-actions">
            <button className="ghost-button" onClick={onRefresh} type="button">
              Проверить
            </button>
            <button className="primary-button" disabled={state.busy || !selected?.id} onClick={onUpdate} type="button">
              {state.busy ? "Обновление..." : selected?.has_update ? "Обновить" : "Обновить сейчас"}
            </button>
            <button className="launch-button" disabled={!selected?.id} onClick={onLaunch} type="button">
              Запустить
            </button>
          </div>
        </section>

        <section className="grid">
          <div className="panel detail-card">
            <span className="eyebrow">Application</span>
            <div className="detail-grid">
              <div>
                <span>Папка</span>
                <strong>{selected?.install_dir ?? "n/a"}</strong>
              </div>
              <div>
                <span>Executable</span>
                <strong>{selected?.executable ?? "n/a"}</strong>
              </div>
              <div>
                <span>Текущая версия</span>
                <strong>{selected?.current_version ?? "n/a"}</strong>
              </div>
              <div>
                <span>Версия на сервере</span>
                <strong>{selected?.latest_version ?? "n/a"}</strong>
              </div>
            </div>
          </div>
          <div className="panel progress-card">
            <div className="progress-card__head">
              <span className="eyebrow">Progress</span>
              <strong>{Math.round(state.progress.percent)}%</strong>
            </div>
            <div className="progress-bar">
              <div className="progress-bar__fill" style={{ width: `${Math.max(0, Math.min(100, state.progress.percent))}%` }} />
            </div>
            <div className="progress-copy">
              <strong>{state.progress.message}</strong>
              <span>{progressLabel}</span>
            </div>
            {state.error ? <div className="error-pill">{state.error}</div> : null}
            {launchError ? <div className="error-pill">{launchError}</div> : null}
          </div>
        </section>

        <section className="logs panel">
          <div className="logs__head">
            <div>
              <span className="eyebrow">Logs</span>
              <h2>Последние события</h2>
            </div>
          </div>
          <div className="logs__list">
            {logs.length === 0 ? <div className="log-line empty">Логи появятся после проверки или обновления.</div> : null}
            {logs.map((entry, index) => (
              <div className={`log-line level-${entry.level.toLowerCase()}`} key={`${entry.time}-${index}`}>
                <span>{entry.time}</span>
                <strong>{entry.level}</strong>
                <p>{entry.message}</p>
              </div>
            ))}
          </div>
        </section>
      </main>
    </div>
  );
}

function statusLabel(status: string): string {
  switch (status) {
    case "update-available":
      return "update";
    case "up-to-date":
      return "ok";
    case "offline":
      return "offline";
    case "missing":
      return "missing";
    case "config":
    case "error":
      return "error";
    default:
      return "ready";
  }
}

function formatBytes(value: number): string {
  if (value <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  let size = value;
  let index = 0;
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024;
    index += 1;
  }
  return `${size.toFixed(index === 0 ? 0 : 1)} ${units[index]}`;
}

function normalizeSnapshot(value: unknown): Snapshot {
  if (!value || typeof value !== "object") return emptyState;
  const raw = value as Partial<Snapshot>;
  return {
    ...emptyState,
    ...raw,
    apps: Array.isArray(raw.apps) ? raw.apps : [],
    logs: Array.isArray(raw.logs) ? raw.logs : [],
    progress: { ...emptyState.progress, ...(raw.progress ?? {}) },
  };
}

export default App;
