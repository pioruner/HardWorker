import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import ReactECharts from "echarts-for-react";
import "./App.scss";
import { ApplyControls, ClearRows, ExportRows, GetLogs, GetSnapshot, SetCursorIndex } from "../wailsjs/go/main/App";

type ViskoRow = {
  t1: number;
  t2: number;
  u1: number;
  u2: number;
  temp: number;
};

type ViskoSnapshot = {
  connected: boolean;
  lastResponse: string;
  address: string;
  rows: ViskoRow[];
  cursorIndex: number;
  curT1: string;
  curT2: string;
  curU1: string;
  curU2: string;
  curTemp: string;
  curCmd: string;
  selT1: string;
  selT2: string;
  selU1: string;
  selU2: string;
  selTemp: string;
};

type LogEntry = {
  time: string;
  level: "INFO" | "WARN" | "ERROR" | "DEBUG";
  message: string;
};

const defaultSnapshot: ViskoSnapshot = {
  connected: false,
  lastResponse: "Ожидание подключения к вискозиметру...",
  address: "192.168.0.200:502",
  rows: [],
  cursorIndex: 0,
  curT1: "0",
  curT2: "0",
  curU1: "0.00",
  curU2: "0.00",
  curTemp: "0.0",
  curCmd: "0",
  selT1: "0",
  selT2: "0",
  selU1: "0.00",
  selU2: "0.00",
  selTemp: "0.0",
};

function asSnapshot(value: unknown): ViskoSnapshot {
  if (!value || typeof value !== "object") {
    return defaultSnapshot;
  }
  const raw = value as Partial<ViskoSnapshot>;
  return {
    ...defaultSnapshot,
    ...raw,
    rows: Array.isArray(raw.rows) ? raw.rows : [],
  };
}

function asLogs(value: unknown): LogEntry[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value
    .filter((it) => it && typeof it === "object")
    .map((it) => {
      const row = it as Partial<LogEntry>;
      return {
        time: row.time ?? "",
        level: row.level ?? "INFO",
        message: row.message ?? "",
      } as LogEntry;
    });
}

function App() {
  const [view, setView] = useState<"visco" | "logs">("visco");
  const [logsPaused, setLogsPaused] = useState(false);
  const [logsUiOffset, setLogsUiOffset] = useState(0);
  const timeChartRef = useRef<ReactECharts>(null);
  const voltageChartRef = useRef<ReactECharts>(null);
  const tempChartRef = useRef<ReactECharts>(null);
  const timeChartHostRef = useRef<HTMLDivElement>(null);
  const voltageChartHostRef = useRef<HTMLDivElement>(null);
  const tempChartHostRef = useRef<HTMLDivElement>(null);
  const stabilizeTimerRef = useRef<number | null>(null);
  const [snapshot, setSnapshot] = useState<ViskoSnapshot>(defaultSnapshot);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [addressDraft, setAddressDraft] = useState(defaultSnapshot.address);

  useEffect(() => {
    setAddressDraft(snapshot.address);
  }, [snapshot.address]);

  useEffect(() => {
    if (logsUiOffset > logs.length) {
      setLogsUiOffset(logs.length);
    }
  }, [logs.length, logsUiOffset]);

  useEffect(() => {
    let active = true;

    const sync = async () => {
      try {
        const [snapshotRaw, logsRaw] = await Promise.all([GetSnapshot(), GetLogs()]);
        if (!active) {
          return;
        }
        setSnapshot(asSnapshot(snapshotRaw));
        if (!logsPaused) {
          setLogs(asLogs(logsRaw));
        }
      } catch {
        // ignore polling errors
      }
    };

    void sync();
    const id = window.setInterval(sync, 500);
    return () => {
      active = false;
      window.clearInterval(id);
    };
  }, [logsPaused]);

  const rows = snapshot.rows;
  const x = useMemo(() => rows.map((_, idx) => idx), [rows]);
  const t1 = useMemo(() => rows.map((r) => r.t1), [rows]);
  const t2 = useMemo(() => rows.map((r) => r.t2), [rows]);
  const u1 = useMemo(() => rows.map((r) => r.u1), [rows]);
  const u2 = useMemo(() => rows.map((r) => r.u2), [rows]);
  const temp = useMemo(() => rows.map((r) => r.temp), [rows]);

  const xMax = x.length > 1 ? x[x.length - 1] : 1;

  const makeChartOptions = (series: Array<{ name: string; data: number[] }>) => ({
    animation: false,
    backgroundColor: "transparent",
    tooltip: { trigger: "axis" },
    legend: {
      textStyle: { color: "#bdc8d8" },
      data: series.map((s) => s.name),
      top: 4,
    },
    grid: {
      top: 34,
      right: 18,
      bottom: 28,
      left: 48,
    },
    xAxis: {
      type: "value",
      min: 0,
      max: xMax,
      axisLabel: { color: "#bdc8d8", formatter: (value: number) => `${Math.round(value)}` },
      splitLine: { lineStyle: { color: "rgba(136, 156, 180, 0.18)" } },
    },
    yAxis: {
      type: "value",
      axisLabel: { color: "#bdc8d8" },
      splitLine: { lineStyle: { color: "rgba(136, 156, 180, 0.18)" } },
    },
    series: series.map((s) => ({
      name: s.name,
      type: "line",
      symbol: "none",
      data: x.map((v, i) => [v, s.data[i]]),
      markLine: {
        symbol: "none",
        label: { show: false },
        data: [{ xAxis: snapshot.cursorIndex, lineStyle: { color: "#ffc857", width: 2 } }],
      },
    })),
  });

  const timeChartOptions = useMemo(
    () => makeChartOptions([{ name: "T1", data: t1 }, { name: "T2", data: t2 }]),
    [snapshot.cursorIndex, t1, t2, xMax],
  );
  const voltageChartOptions = useMemo(
    () => makeChartOptions([{ name: "U1", data: u1 }, { name: "U2", data: u2 }]),
    [snapshot.cursorIndex, u1, u2, xMax],
  );
  const tempChartOptions = useMemo(
    () => makeChartOptions([{ name: "Temp", data: temp }]),
    [snapshot.cursorIndex, temp, xMax],
  );

  const onApplyAddress = async () => {
    try {
      const next = await ApplyControls({ address: addressDraft });
      setSnapshot(asSnapshot(next));
    } catch {
      // ignore backend availability race
    }
  };

  const onCursorChange = async (idx: number) => {
    try {
      const next = await SetCursorIndex(idx);
      setSnapshot(asSnapshot(next));
    } catch {
      // ignore transient errors
    }
  };

  const onClearRows = async () => {
    try {
      const next = await ClearRows();
      setSnapshot(asSnapshot(next));
    } catch {
      // ignore transient errors
    }
  };

  const onExportRows = async () => {
    try {
      const next = await ExportRows();
      setSnapshot(asSnapshot(next));
    } catch {
      // ignore transient errors
    }
  };

  const visibleLogs = logs.slice(logsUiOffset);

  const resizeChart = useCallback((host: HTMLDivElement | null, chartRef: React.RefObject<ReactECharts | null>) => {
    const chart = chartRef.current?.getEchartsInstance();
    if (!host || !chart) {
      return;
    }
    const width = host.clientWidth;
    const height = host.clientHeight;
    if (width <= 0 || height <= 0) {
      return;
    }
    chart.resize({ width, height });
  }, []);

  const forceChartsResize = useCallback(() => {
    resizeChart(timeChartHostRef.current, timeChartRef);
    resizeChart(voltageChartHostRef.current, voltageChartRef);
    resizeChart(tempChartHostRef.current, tempChartRef);
  }, [resizeChart]);

  const startChartStabilization = useCallback(() => {
    if (stabilizeTimerRef.current !== null) {
      window.clearInterval(stabilizeTimerRef.current);
      stabilizeTimerRef.current = null;
    }
    let ticks = 0;
    stabilizeTimerRef.current = window.setInterval(() => {
      forceChartsResize();
      ticks += 1;
      if (ticks >= 20) {
        if (stabilizeTimerRef.current !== null) {
          window.clearInterval(stabilizeTimerRef.current);
          stabilizeTimerRef.current = null;
        }
      }
    }, 100);
  }, [forceChartsResize]);

  useEffect(() => {
    const hosts = [timeChartHostRef.current, voltageChartHostRef.current, tempChartHostRef.current].filter(
      (host): host is HTMLDivElement => host !== null,
    );
    if (hosts.length === 0) {
      return;
    }

    const observer = new ResizeObserver(() => {
      forceChartsResize();
    });
    hosts.forEach((host) => observer.observe(host));

    return () => {
      observer.disconnect();
      if (stabilizeTimerRef.current !== null) {
        window.clearInterval(stabilizeTimerRef.current);
        stabilizeTimerRef.current = null;
      }
    };
  }, [forceChartsResize]);

  useEffect(() => {
    if (view !== "visco") {
      return;
    }
    const resize = () => forceChartsResize();
    const raf = window.requestAnimationFrame(resize);
    const t1 = window.setTimeout(resize, 120);
    const t2 = window.setTimeout(resize, 380);
    startChartStabilization();
    return () => {
      window.cancelAnimationFrame(raf);
      window.clearTimeout(t1);
      window.clearTimeout(t2);
      if (stabilizeTimerRef.current !== null) {
        window.clearInterval(stabilizeTimerRef.current);
        stabilizeTimerRef.current = null;
      }
    };
  }, [forceChartsResize, startChartStabilization, view]);

  return (
    <main className="akip-layout">
      <header className="panel panel-mode">
        <button className={view === "visco" ? "is-active" : ""} onClick={() => setView("visco")}>
          VISCO
        </button>
        <button className={view === "logs" ? "is-active" : ""} onClick={() => setView("logs")}>
          Логи
        </button>
      </header>

      {view === "logs" ? (
        <section className="panel logs-panel">
          <div className="logs-head">
            <div className="panel-title">Системный лог</div>
            <div className="logs-actions">
              <button className={`toggle ${logsPaused ? "is-active" : ""}`} onClick={() => setLogsPaused((v) => !v)}>
                {logsPaused ? "Продолжить" : "Пауза"}
              </button>
              <button className="toggle" onClick={() => setLogsUiOffset(logs.length)}>
                Очистить
              </button>
            </div>
          </div>
          <div className="logs-list">
            {[...visibleLogs].reverse().map((line, idx) => (
              <div className={`log-line level-${line.level.toLowerCase()}`} key={`${line.time}-${idx}`}>
                <span className="log-time">{line.time}</span>
                <span className="log-level">{line.level}</span>
                <span className="log-msg">{line.message}</span>
              </div>
            ))}
          </div>
        </section>
      ) : (
        <div className="akip-view visco-view">
          <section className="visco-main">
            <aside className="panel visco-sidebar">
              <header className="visco-sidebar-top">
                <label className="field">
                  <span>Адрес прибора</span>
                  <input value={addressDraft} onChange={(e) => setAddressDraft(e.target.value)} onBlur={onApplyAddress} />
                </label>
                <div className={`status ${snapshot.connected ? "is-online" : "is-offline"}`}>
                  {snapshot.connected ? "Связь: онлайн" : "Связь: оффлайн"}
                </div>
                <div className="visco-sidebar-actions">
                  <button className="toggle" onClick={onExportRows}>Сохранить CSV</button>
                  <button className="toggle" onClick={onClearRows}>Очистить</button>
                </div>
              </header>

              <section className="panel panel-metrics visco-metrics">
                <div className="metric"><span>T1</span><strong>{snapshot.curT1}</strong></div>
                <div className="metric"><span>T2</span><strong>{snapshot.curT2}</strong></div>
                <div className="metric"><span>U1</span><strong>{snapshot.curU1}</strong></div>
                <div className="metric"><span>U2</span><strong>{snapshot.curU2}</strong></div>
                <div className="metric"><span>Temp</span><strong>{snapshot.curTemp}</strong></div>
                <div className="metric"><span>CMD</span><strong>{snapshot.curCmd}</strong></div>
                <div className="response-row">
                  <span>Последний ответ</span>
                  <input className="response-box" readOnly value={snapshot.lastResponse} title={snapshot.lastResponse} />
                </div>
              </section>

              <section className="visco-table-panel">
                <div className="panel-title">Таблица измерений</div>
                <div className="visco-table-wrap">
                  <table className="visco-table">
                    <thead>
                      <tr>
                        <th>#</th>
                        <th>T1</th>
                        <th>T2</th>
                        <th>U1</th>
                        <th>U2</th>
                        <th>Temp</th>
                      </tr>
                    </thead>
                    <tbody>
                      {rows.map((r, i) => (
                        <tr key={i} className={i === snapshot.cursorIndex ? "is-active" : ""}>
                          <td>{i}</td>
                          <td>{r.t1.toFixed(0)}</td>
                          <td>{r.t2.toFixed(0)}</td>
                          <td>{r.u1.toFixed(2)}</td>
                          <td>{r.u2.toFixed(2)}</td>
                          <td>{r.temp.toFixed(2)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </section>
            </aside>

            <section className="panel visco-workspace">
              <div className="panel-title">Тренды VISCO</div>
              <div className="visco-charts">
                <div className="chart-host visco-chart" ref={timeChartHostRef}>
                  <ReactECharts ref={timeChartRef} option={timeChartOptions} style={{ height: "100%", width: "100%" }} />
                </div>
                <div className="chart-host visco-chart" ref={voltageChartHostRef}>
                  <ReactECharts ref={voltageChartRef} option={voltageChartOptions} style={{ height: "100%", width: "100%" }} />
                </div>
                <div className="chart-host visco-chart" ref={tempChartHostRef}>
                  <ReactECharts ref={tempChartRef} option={tempChartOptions} style={{ height: "100%", width: "100%" }} />
                </div>
              </div>

              <section className="panel panel-bottom visco-bottom">
                <label className="slider-wrap">
                  <span>Курсор записи</span>
                  <input
                    type="range"
                    min={0}
                    max={Math.max(0, rows.length - 1)}
                    step={1}
                    value={Math.min(snapshot.cursorIndex, Math.max(0, rows.length - 1))}
                    onChange={(e) => void onCursorChange(Number(e.target.value))}
                    disabled={rows.length === 0}
                  />
                  <b>{snapshot.cursorIndex}</b>
                </label>

                <div className="metric"><span>T1 (cursor)</span><strong>{snapshot.selT1}</strong></div>
                <div className="metric"><span>T2 (cursor)</span><strong>{snapshot.selT2}</strong></div>
                <div className="metric"><span>U1 (cursor)</span><strong>{snapshot.selU1}</strong></div>
                <div className="metric"><span>U2 (cursor)</span><strong>{snapshot.selU2}</strong></div>
                <div className="metric"><span>Temp (cursor)</span><strong>{snapshot.selTemp}</strong></div>
              </section>
            </section>
          </section>
        </div>
      )}
    </main>
  );
}

export default App;
