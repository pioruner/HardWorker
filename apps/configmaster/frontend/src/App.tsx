import { useEffect, useMemo, useState } from "react";
import "./App.scss";
import { GenerateConfigs, GetState, SaveMasterSettings, SelectDirectory } from "../wailsjs/go/main/App";
import logoUrl from "./assets/images/logo-universal.png";

type ManagedAppConfig = {
  id: string;
  label: string;
  platform: string;
  arch: string;
  install_dir: string;
  executable: string;
};

type MasterSettings = {
  project_name: string;
  manifest_url: string;
  s3_endpoint: string;
  s3_bucket: string;
  s3_region: string;
  s3_tenant_id: string;
  s3_access_key_id: string;
  s3_secret_key: string;
  s3_prefix: string;
  s3_public_base_url: string;
  uploader_config_path: string;
  updater_config_path: string;
  apps: ManagedAppConfig[];
};

type LogEntry = {
  time: string;
  level: string;
  message: string;
};

type MasterSnapshot = {
  config_path: string;
  ready: boolean;
  error: string;
  settings: MasterSettings;
  logs: LogEntry[];
};

const emptySettings: MasterSettings = {
  project_name: "",
  manifest_url: "",
  s3_endpoint: "",
  s3_bucket: "",
  s3_region: "ru-central-1",
  s3_tenant_id: "",
  s3_access_key_id: "",
  s3_secret_key: "",
  s3_prefix: "",
  s3_public_base_url: "",
  uploader_config_path: "config/uploader.local.json",
  updater_config_path: "config/updater.local.json",
  apps: [
    {
      id: "app",
      label: "Application",
      platform: "windows",
      arch: "amd64",
      install_dir: "./app",
      executable: "app.exe",
    },
  ],
};

const emptyState: MasterSnapshot = {
  config_path: "",
  ready: false,
  error: "",
  settings: emptySettings,
  logs: [],
};

function App() {
  const [state, setState] = useState<MasterSnapshot>(emptyState);
  const [form, setForm] = useState<MasterSettings>(emptySettings);
  const [dirty, setDirty] = useState(false);
  const [busy, setBusy] = useState(false);
  const [activeAppID, setActiveAppID] = useState("app");

  useEffect(() => {
    let active = true;
    const sync = async () => {
      try {
        const next = normalizeSnapshot(await GetState());
        if (!active) return;
        setState(next);
        setForm((current) => (dirty || busy ? current : next.settings));
        if (!activeAppID && next.settings.apps.length > 0) {
          setActiveAppID(next.settings.apps[0].id);
        }
      } catch {
        // ignore polling hiccups
      }
    };
    void sync();
    const id = window.setInterval(sync, 1200);
    return () => {
      active = false;
      window.clearInterval(id);
    };
  }, [activeAppID, dirty, busy]);

  const selectedApp = useMemo(
    () => form.apps.find((app) => app.id === activeAppID) ?? form.apps[0],
    [activeAppID, form.apps],
  );

  const patchSettings = <K extends keyof MasterSettings>(key: K, value: MasterSettings[K]) => {
    setDirty(true);
    setForm((current) => ({ ...current, [key]: value }));
  };

  const patchApp = (patch: Partial<ManagedAppConfig>) => {
    setDirty(true);
    setForm((current) => ({
      ...current,
      apps: current.apps.map((app) => (app.id === activeAppID ? { ...app, ...patch } : app)),
    }));
  };

  const addApp = () => {
    const id = `app${form.apps.length + 1}`;
    const nextApp: ManagedAppConfig = {
      id,
      label: `Application ${form.apps.length + 1}`,
      platform: "windows",
      arch: "amd64",
      install_dir: `./${id}`,
      executable: `${id}.exe`,
    };
    setDirty(true);
    setForm((current) => ({ ...current, apps: [...current.apps, nextApp] }));
    setActiveAppID(id);
  };

  const removeApp = () => {
    if (form.apps.length <= 1) return;
    const nextApps = form.apps.filter((app) => app.id !== activeAppID);
    setDirty(true);
    setForm((current) => ({ ...current, apps: nextApps }));
    setActiveAppID(nextApps[0]?.id ?? "");
  };

  const pickDirectory = async (target: "install" | "uploader" | "updater") => {
    const path = await SelectDirectory();
    if (!path) return;
    if (target === "install") patchApp({ install_dir: path });
    if (target === "uploader") patchSettings("uploader_config_path", `${path}/uploader.local.json`);
    if (target === "updater") patchSettings("updater_config_path", `${path}/updater.local.json`);
  };

  const onSave = async () => {
    setBusy(true);
    try {
      const next = normalizeSnapshot(await SaveMasterSettings(form as any));
      setState(next);
      setForm(next.settings);
      setDirty(false);
    } finally {
      setBusy(false);
    }
  };

  const onGenerate = async () => {
    setBusy(true);
    try {
      const next = normalizeSnapshot(await GenerateConfigs(form as any));
      setState(next);
      setForm(next.settings);
      setDirty(false);
    } finally {
      setBusy(false);
    }
  };

  const logs = state.logs.slice(0, 12);

  return (
    <div className="updater-layout">
      <aside className="sidebar panel">
        <div className="brand">
          <img alt="ConfigMaster" src={logoUrl} />
          <div>
            <strong>HardWorker ConfigMaster</strong>
            <span>генерация uploader/updater конфигов</span>
          </div>
        </div>
        <div className="config-card">
          <span>Master Config</span>
          <code>{state.config_path || "config/configmaster.local.json"}</code>
        </div>
        <button className="ghost-button" onClick={addApp} type="button">
          Добавить приложение
        </button>
        <div className="app-list">
          {form.apps.map((app) => (
            <button
              key={app.id}
              className={`app-card ${activeAppID === app.id ? "is-active" : ""}`}
              onClick={() => setActiveAppID(app.id)}
              type="button"
            >
              <div className="app-card__head">
                <strong>{app.label || app.id}</strong>
                <span className="status status--ready">{app.platform}/{app.arch}</span>
              </div>
              <div className="app-card__meta">
                <span>ID</span>
                <strong>{app.id}</strong>
              </div>
              <div className="app-card__meta">
                <span>Exe</span>
                <strong>{app.executable}</strong>
              </div>
            </button>
          ))}
        </div>
      </aside>

      <main className="content">
        <section className="hero panel">
          <div>
            <span className="eyebrow">Master Setup</span>
            <h1>{form.project_name || "Новый универсальный проект"}</h1>
            <p>{state.error || "Сначала сохраняется мастер-конфиг, затем ConfigMaster генерирует uploader.local.json и updater.local.json."}</p>
          </div>
          <div className="hero-actions">
            <button className="ghost-button" onClick={onSave} type="button">
              {busy ? "Сохранение..." : dirty ? "Сохранить мастер*" : "Сохранить мастер"}
            </button>
            <button className="primary-button" onClick={onGenerate} type="button">
              Сгенерировать файлы
            </button>
            <button className="launch-button" disabled={form.apps.length <= 1} onClick={removeApp} type="button">
              Удалить приложение
            </button>
          </div>
        </section>

        <section className="grid grid--settings">
          <div className="panel detail-card">
            <span className="eyebrow">Project</span>
            <div className="form-grid">
              <label className="field field--wide">
                <span>Project Name</span>
                <input value={form.project_name} onChange={(event) => patchSettings("project_name", event.target.value)} />
              </label>
              <label className="field field--wide">
                <span>Manifest URL</span>
                <input value={form.manifest_url} onChange={(event) => patchSettings("manifest_url", event.target.value)} />
              </label>
              <label className="field field--wide">
                <span>Uploader Config Path</span>
                <div className="input-row">
                  <input value={form.uploader_config_path} onChange={(event) => patchSettings("uploader_config_path", event.target.value)} />
                  <button className="ghost-button ghost-button--small" onClick={() => pickDirectory("uploader")} type="button">
                    Папка
                  </button>
                </div>
              </label>
              <label className="field field--wide">
                <span>Updater Config Path</span>
                <div className="input-row">
                  <input value={form.updater_config_path} onChange={(event) => patchSettings("updater_config_path", event.target.value)} />
                  <button className="ghost-button ghost-button--small" onClick={() => pickDirectory("updater")} type="button">
                    Папка
                  </button>
                </div>
              </label>
            </div>
          </div>

          <div className="panel detail-card">
            <span className="eyebrow">Storage</span>
            <div className="form-grid">
              <label className="field field--wide">
                <span>S3 Endpoint</span>
                <input value={form.s3_endpoint} onChange={(event) => patchSettings("s3_endpoint", event.target.value)} />
              </label>
              <label className="field">
                <span>Bucket</span>
                <input value={form.s3_bucket} onChange={(event) => patchSettings("s3_bucket", event.target.value)} />
              </label>
              <label className="field">
                <span>Region</span>
                <input value={form.s3_region} onChange={(event) => patchSettings("s3_region", event.target.value)} />
              </label>
              <label className="field field--wide">
                <span>Tenant ID</span>
                <input value={form.s3_tenant_id} onChange={(event) => patchSettings("s3_tenant_id", event.target.value)} />
              </label>
              <label className="field">
                <span>Access Key</span>
                <input value={form.s3_access_key_id} onChange={(event) => patchSettings("s3_access_key_id", event.target.value)} />
              </label>
              <label className="field">
                <span>Secret Key</span>
                <input type="password" value={form.s3_secret_key} onChange={(event) => patchSettings("s3_secret_key", event.target.value)} />
              </label>
              <label className="field">
                <span>Prefix</span>
                <input value={form.s3_prefix} onChange={(event) => patchSettings("s3_prefix", event.target.value)} />
              </label>
              <label className="field">
                <span>Public Base URL</span>
                <input value={form.s3_public_base_url} onChange={(event) => patchSettings("s3_public_base_url", event.target.value)} />
              </label>
            </div>
          </div>
        </section>

        <section className="grid">
          <div className="panel detail-card">
            <span className="eyebrow">Selected App</span>
            <div className="form-grid">
              <label className="field">
                <span>ID</span>
                <input value={selectedApp?.id ?? ""} onChange={(event) => patchApp({ id: event.target.value })} />
              </label>
              <label className="field">
                <span>Label</span>
                <input value={selectedApp?.label ?? ""} onChange={(event) => patchApp({ label: event.target.value })} />
              </label>
              <label className="field">
                <span>Platform</span>
                <input value={selectedApp?.platform ?? ""} onChange={(event) => patchApp({ platform: event.target.value })} />
              </label>
              <label className="field">
                <span>Arch</span>
                <input value={selectedApp?.arch ?? ""} onChange={(event) => patchApp({ arch: event.target.value })} />
              </label>
              <label className="field field--wide">
                <span>Install Dir</span>
                <div className="input-row">
                  <input value={selectedApp?.install_dir ?? ""} onChange={(event) => patchApp({ install_dir: event.target.value })} />
                  <button className="ghost-button ghost-button--small" onClick={() => pickDirectory("install")} type="button">
                    Папка
                  </button>
                </div>
              </label>
              <label className="field field--wide">
                <span>Executable</span>
                <input value={selectedApp?.executable ?? ""} onChange={(event) => patchApp({ executable: event.target.value })} />
              </label>
            </div>
          </div>

          <div className="panel logs">
            <div className="logs__head">
              <div>
                <span className="eyebrow">Logs</span>
                <h2>Последние события</h2>
              </div>
            </div>
            <div className="logs__list">
              {logs.length === 0 ? <div className="log-line empty">Логи появятся после сохранения или генерации файлов.</div> : null}
              {logs.map((entry, index) => (
                <div className={`log-line level-${entry.level.toLowerCase()}`} key={`${entry.time}-${index}`}>
                  <span>{entry.time}</span>
                  <strong>{entry.level}</strong>
                  <p>{entry.message}</p>
                </div>
              ))}
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}

function normalizeSnapshot(value: unknown): MasterSnapshot {
  if (!value || typeof value !== "object") return emptyState;
  const raw = value as Partial<MasterSnapshot>;
  return {
    ...emptyState,
    ...raw,
    logs: Array.isArray(raw.logs) ? raw.logs : [],
    settings: {
      ...emptySettings,
      ...(raw.settings ?? {}),
      apps: Array.isArray(raw.settings?.apps) ? raw.settings?.apps ?? emptySettings.apps : emptySettings.apps,
    },
  };
}

export default App;
