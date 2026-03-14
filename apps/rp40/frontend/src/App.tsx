import { useEffect, useMemo, useRef, useState } from "react";
import ReactECharts from "echarts-for-react";
import "./App.scss";
import { Calculate, GetSnapshot, SelectMeasurementFile, SelectPassportFile, SetSelectedSample, UpdateInputs } from "../wailsjs/go/main/App";

type LogEntry = {
  Time?: string;
  Level?: string;
  Message?: string;
};

type PassportSampleOption = {
  ID?: string;
  Label?: string;
  Occurrences?: number;
  UpdatedAt?: string;
};

type PassportSampleRecord = {
  SampleID?: string;
  Timestamp?: string;
  Gas?: string;
  LengthMM?: number;
  DiameterMM?: number;
  HeightMM?: number;
  VolumeCM3?: number;
  PorosityPct?: number;
  TemperatureC?: number;
  AtmosphericKPa?: number;
  CompressionMPa?: number;
  PoreVolumeCM3?: number;
  KGasMD?: number;
  KKlinkenbergMD?: number;
  BPsi?: number;
  RawRow?: number;
  SourceOccurrences?: number;
};

type RP40Inputs = {
  SampleID?: string;
  Gas?: string;
  DiameterMM?: number;
  LengthMM?: number;
  PorosityPct?: number;
  TemperatureC?: number;
  AtmosphericKPa?: number;
  ReservoirMode?: string;
  ReservoirVolumeML?: number;
  CustomReservoirML?: number;
  LowPerm?: boolean;
  MinDeltaP?: number;
  MinDeltaT?: number;
  StartAt?: number;
};

type ChartPoint = { X?: number; Y?: number };
type B22LinearityPoint = {
  MeanPressurePa?: number;
  MeanPressurePsi?: number;
  Response?: number;
  FitResponse?: number;
  Residual?: number;
};

type CalculationView = {
  HasResult?: boolean;
  ProcessedRows?: number;
  DisplayKMD?: number;
  DisplayBPsi?: number;
  DisplayMethod?: string;
  DisplayWarnings?: string[];
  ProcessedCurve?: ChartPoint[];
  B22FitCurve?: ChartPoint[];
  Full?: {
    KMD?: number;
    BPa?: number;
    BPsi?: number;
    SE?: number;
    R2?: number;
    A1?: number;
    A2?: number;
  };
  B22?: {
    KMD?: number;
    BPsi?: number;
    R2?: number;
    Slope?: number;
    Intercept?: number;
    MaxResidual?: number;
    Linearity?: B22LinearityPoint[];
  };
  Assessment?: {
    Recommendation?: string;
    KSource?: string;
    BSource?: string;
    RecommendedKMD?: number;
    RecommendedBPsi?: number;
    Rationale?: string[];
    B22Valid?: boolean;
    B22Marginal?: boolean;
    FullBCollapsed?: boolean;
    HighPerm?: boolean;
    RelativeKGap?: number;
    MeanPmPsi?: number;
  };
};

type Snapshot = {
  PassportFilePath?: string;
  MeasurementFilePath?: string;
  Samples?: PassportSampleOption[];
  SelectedSample?: PassportSampleRecord;
  Inputs?: RP40Inputs;
  Calculation?: CalculationView;
  Logs?: LogEntry[];
  LastError?: string;
};

const defaultSnapshot: Snapshot = {
  Samples: [],
  SelectedSample: {},
  Inputs: {
    ReservoirMode: "233",
    ReservoirVolumeML: 233,
    StartAt: 0.85,
  },
  Calculation: {
    HasResult: false,
    DisplayWarnings: [],
    ProcessedCurve: [],
    B22FitCurve: [],
    Full: {},
    B22: {
      Linearity: [],
    },
    Assessment: {
      Rationale: [],
    },
  },
  Logs: [],
  LastError: "",
};

function asSnapshot(raw: unknown): Snapshot {
  if (!raw || typeof raw !== "object") {
    return defaultSnapshot;
  }
  const next = raw as Snapshot;
  return {
    ...defaultSnapshot,
    ...next,
    Samples: next.Samples ?? [],
    SelectedSample: next.SelectedSample ?? {},
    Inputs: { ...defaultSnapshot.Inputs, ...(next.Inputs ?? {}) },
    Calculation: {
      ...defaultSnapshot.Calculation,
      ...(next.Calculation ?? {}),
      Full: { ...defaultSnapshot.Calculation?.Full, ...(next.Calculation?.Full ?? {}) },
      B22: {
        ...defaultSnapshot.Calculation?.B22,
        ...(next.Calculation?.B22 ?? {}),
        Linearity: next.Calculation?.B22?.Linearity ?? [],
      },
      Assessment: {
        ...defaultSnapshot.Calculation?.Assessment,
        ...(next.Calculation?.Assessment ?? {}),
        Rationale: next.Calculation?.Assessment?.Rationale ?? [],
      },
      DisplayWarnings: next.Calculation?.DisplayWarnings ?? [],
      ProcessedCurve: next.Calculation?.ProcessedCurve ?? [],
      B22FitCurve: next.Calculation?.B22FitCurve ?? [],
    },
    Logs: next.Logs ?? [],
    LastError: next.LastError ?? "",
  };
}

function formatNumber(value: number | undefined, digits = 4): string {
  if (typeof value !== "number" || Number.isNaN(value)) {
    return "—";
  }
  return value.toFixed(digits);
}

function toInputValue(value: number | undefined): string {
  if (typeof value !== "number" || Number.isNaN(value)) {
    return "";
  }
  return String(value);
}

function methodLabel(value: string | undefined): string {
  switch (value) {
    case "b22":
      return "B-22";
    case "full":
      return "Full-fit";
    case "hybrid":
      return "Гибрид";
    case "review":
      return "Нужна проверка";
    default:
      return value || "—";
  }
}

function sourceLabel(value: string | undefined): string {
  switch (value) {
    case "full":
      return "Full-fit";
    case "review":
      return "Проверка";
    default:
      return value || "—";
  }
}

function App() {
  const [snapshot, setSnapshot] = useState<Snapshot>(defaultSnapshot);
  const [busy, setBusy] = useState(false);
  const [view, setView] = useState<"results" | "logs">("results");
  const [showAdvanced, setShowAdvanced] = useState(false);
  const processedChartRef = useRef<ReactECharts>(null);
  const b22ChartRef = useRef<ReactECharts>(null);
  const processedHostRef = useRef<HTMLDivElement>(null);
  const b22HostRef = useRef<HTMLDivElement>(null);

  const inputs = snapshot.Inputs ?? defaultSnapshot.Inputs!;
  const calculation = snapshot.Calculation ?? defaultSnapshot.Calculation!;
  const selectedSample = snapshot.SelectedSample ?? {};
  const samples = snapshot.Samples ?? [];
  const logs = snapshot.Logs ?? [];

  const refresh = async (request: Promise<unknown>) => {
    setBusy(true);
    try {
      setSnapshot(asSnapshot(await request));
    } finally {
      setBusy(false);
    }
  };

  useEffect(() => {
    void refresh(GetSnapshot());
  }, []);

  useEffect(() => {
    if (view !== "results") {
      return;
    }

    const resizeCharts = () => {
      const pairs = [
        { host: processedHostRef.current, chart: processedChartRef.current?.getEchartsInstance() },
        { host: b22HostRef.current, chart: b22ChartRef.current?.getEchartsInstance() },
      ];
      pairs.forEach(({ host, chart }) => {
        if (!host || !chart) {
          return;
        }
        const width = host.clientWidth;
        const height = host.clientHeight;
        if (width > 0 && height > 0) {
          chart.resize({ width, height });
        }
      });
    };

    const rafId = window.requestAnimationFrame(resizeCharts);
    const timers = [120, 280, 520].map((delay) => window.setTimeout(resizeCharts, delay));
    const observer = new ResizeObserver(() => resizeCharts());
    if (processedHostRef.current) {
      observer.observe(processedHostRef.current);
    }
    if (b22HostRef.current) {
      observer.observe(b22HostRef.current);
    }
    window.addEventListener("resize", resizeCharts);

    return () => {
      window.cancelAnimationFrame(rafId);
      timers.forEach((timer) => window.clearTimeout(timer));
      observer.disconnect();
      window.removeEventListener("resize", resizeCharts);
    };
  }, [view, calculation.ProcessedRows, calculation.DisplayMethod]);

  const patchInputs = async (patch: Partial<RP40Inputs>) => {
    const next: RP40Inputs = {
      SampleID: inputs.SampleID ?? "",
      Gas: inputs.Gas ?? "",
      DiameterMM: inputs.DiameterMM ?? 0,
      LengthMM: inputs.LengthMM ?? 0,
      PorosityPct: inputs.PorosityPct ?? 0,
      TemperatureC: inputs.TemperatureC ?? 0,
      AtmosphericKPa: inputs.AtmosphericKPa ?? 0,
      ReservoirMode: inputs.ReservoirMode ?? "233",
      ReservoirVolumeML: inputs.ReservoirVolumeML ?? 233,
      CustomReservoirML: inputs.CustomReservoirML ?? 0,
      LowPerm: inputs.LowPerm ?? false,
      MinDeltaP: inputs.MinDeltaP ?? 0,
      MinDeltaT: inputs.MinDeltaT ?? 0,
      StartAt: inputs.StartAt ?? 0.85,
      ...patch,
    };
    await refresh(UpdateInputs(next as any));
  };

  const processedSeries = useMemo(
    () => (calculation.ProcessedCurve ?? []).map((point) => [point.X ?? 0, point.Y ?? 0]),
    [calculation.ProcessedCurve],
  );

  const b22Points = useMemo(
    () =>
      (calculation.B22?.Linearity ?? []).map((point) => ({
        value: [point.MeanPressurePsi ?? 0, point.Response ?? 0],
      })),
    [calculation.B22?.Linearity],
  );

  const b22Fit = useMemo(
    () => (calculation.B22FitCurve ?? []).map((point) => [point.X ?? 0, point.Y ?? 0]),
    [calculation.B22FitCurve],
  );

  const processedChartOptions = useMemo(
    () => ({
      animation: false,
      backgroundColor: "transparent",
      grid: { left: 54, right: 24, top: 24, bottom: 44 },
      tooltip: { trigger: "axis" },
      xAxis: {
        type: "value",
        name: "t, c",
        axisLabel: { color: "#d9d3c4" },
        splitLine: { lineStyle: { color: "rgba(255,255,255,0.08)" } },
      },
      yAxis: {
        type: "value",
        name: "P, MPa",
        axisLabel: { color: "#d9d3c4" },
        splitLine: { lineStyle: { color: "rgba(255,255,255,0.08)" } },
      },
      series: [{
        type: "line",
        data: processedSeries,
        symbol: "none",
        lineStyle: { width: 2, color: "#eab95f" },
        areaStyle: { color: "rgba(234, 185, 95, 0.12)" },
      }],
    }),
    [processedSeries],
  );

  const b22ChartOptions = useMemo(
    () => ({
      animation: false,
      backgroundColor: "transparent",
      grid: { left: 54, right: 24, top: 24, bottom: 44 },
      tooltip: { trigger: "axis" },
      xAxis: {
        type: "value",
        name: "Pm, psi",
        axisLabel: { color: "#d9d3c4" },
        splitLine: { lineStyle: { color: "rgba(255,255,255,0.08)" } },
      },
      yAxis: {
        type: "value",
        name: "Response",
        axisLabel: { color: "#d9d3c4" },
        splitLine: { lineStyle: { color: "rgba(255,255,255,0.08)" } },
      },
      series: [
        {
          type: "scatter",
          data: b22Points,
          itemStyle: { color: "#72d8ff" },
          symbolSize: 8,
        },
        {
          type: "line",
          data: b22Fit,
          symbol: "none",
          lineStyle: { width: 2, color: "#c6f26d" },
        },
      ],
    }),
    [b22Fit, b22Points],
  );

  return (
    <div className="rp40-layout">
      <header className="hero panel">
        <div>
          <p className="eyebrow">HardWorker / RP40</p>
          <h1>Калькулятор газопроницаемости</h1>
          <p className="hero-copy">
            Загрузите паспорт, загрузите замер, выберите образец и объём Vt.
          </p>
        </div>
        <div className="hero-actions">
          <button onClick={() => void refresh(SelectPassportFile())} disabled={busy}>Загрузить паспорт</button>
          <button onClick={() => void refresh(SelectMeasurementFile())} disabled={busy}>Загрузить замер</button>
          <button className="accent" onClick={() => void refresh(Calculate())} disabled={busy}>Рассчитать</button>
        </div>
      </header>

      <main className="rp40-main">
        <section className="left-column">
          <div className="panel">
            <div className="section-title">Источники</div>
            <div className="path-grid">
              <div>
                <span>Паспорт</span>
                <strong>{snapshot.PassportFilePath || "Не загружен"}</strong>
              </div>
              <div>
                <span>Файл замера</span>
                <strong>{snapshot.MeasurementFilePath || "Не загружен"}</strong>
              </div>
            </div>
            {snapshot.LastError ? <div className="error-box">{snapshot.LastError}</div> : null}
          </div>

          <div className="panel">
            <div className="section-title">Образец и параметры</div>
            <div className="form-grid">
              <label className="field wide">
                <span>Образец</span>
                <select
                  value={selectedSample.SampleID ?? ""}
                  onChange={(event) => void refresh(SetSelectedSample(event.target.value))}
                >
                  <option value="">Выберите образец</option>
                  {samples.map((sample) => (
                    <option key={sample.ID} value={sample.ID}>
                      {sample.Label} {sample.Occurrences && sample.Occurrences > 1 ? `(${sample.Occurrences})` : ""}
                    </option>
                  ))}
                </select>
              </label>
              <label className="field">
                <span>Газ</span>
                <input value={inputs.Gas ?? ""} onChange={(event) => void patchInputs({ Gas: event.target.value })} />
              </label>
              <label className="field">
                <span>Длина, мм</span>
                <input
                  value={toInputValue(inputs.LengthMM)}
                  onChange={(event) => void patchInputs({ LengthMM: Number(event.target.value) })}
                />
              </label>
              <label className="field">
                <span>Диаметр, мм</span>
                <input
                  value={toInputValue(inputs.DiameterMM)}
                  onChange={(event) => void patchInputs({ DiameterMM: Number(event.target.value) })}
                />
              </label>
              <label className="field">
                <span>Пористость, %</span>
                <input
                  value={toInputValue(inputs.PorosityPct)}
                  onChange={(event) => void patchInputs({ PorosityPct: Number(event.target.value) })}
                />
              </label>
              <label className="field">
                <span>Температура, °C</span>
                <input
                  value={toInputValue(inputs.TemperatureC)}
                  onChange={(event) => void patchInputs({ TemperatureC: Number(event.target.value) })}
                />
              </label>
              <label className="field">
                <span>Pатм, кПа</span>
                <input
                  value={toInputValue(inputs.AtmosphericKPa)}
                  onChange={(event) => void patchInputs({ AtmosphericKPa: Number(event.target.value) })}
                />
              </label>
              <label className="field">
                <span>Vt</span>
                <select
                  value={inputs.ReservoirMode ?? "233"}
                  onChange={(event) =>
                    void patchInputs({
                      ReservoirMode: event.target.value,
                      ReservoirVolumeML: event.target.value === "custom" ? inputs.CustomReservoirML : Number(event.target.value),
                    })
                  }
                >
                  <option value="17.3">17,3</option>
                  <option value="233">233</option>
                  <option value="1445">1445</option>
                  <option value="custom">Custom</option>
                </select>
              </label>
              {inputs.ReservoirMode === "custom" && (
                <label className="field">
                  <span>Custom Vt, мл</span>
                  <input
                    value={toInputValue(inputs.CustomReservoirML)}
                    onChange={(event) => void patchInputs({ CustomReservoirML: Number(event.target.value) })}
                  />
                </label>
              )}
            </div>
            <div className="advanced-panel">
              <button
                type="button"
                className="advanced-toggle"
                onClick={() => setShowAdvanced((current) => !current)}
              >
                {showAdvanced ? "Скрыть тонкую настройку" : "Показать тонкую настройку"}
              </button>
              {showAdvanced ? (
                <div className="advanced-grid">
                  <label className="field field-checkbox">
                    <span>LowPerm режим</span>
                    <input
                      type="checkbox"
                      checked={Boolean(inputs.LowPerm)}
                      onChange={(event) => void patchInputs({ LowPerm: event.target.checked })}
                    />
                  </label>
                  <label className="field">
                    <span>Min dP, Pa</span>
                    <input
                      value={toInputValue(inputs.MinDeltaP)}
                      onChange={(event) => void patchInputs({ MinDeltaP: Number(event.target.value) })}
                    />
                  </label>
                  <label className="field">
                    <span>Min dT, s</span>
                    <input
                      value={toInputValue(inputs.MinDeltaT)}
                      onChange={(event) => void patchInputs({ MinDeltaT: Number(event.target.value) })}
                    />
                  </label>
                  <label className="field">
                    <span>StartAt (0..1)</span>
                    <input
                      value={toInputValue(inputs.StartAt)}
                      onChange={(event) => void patchInputs({ StartAt: Number(event.target.value) })}
                    />
                  </label>
                </div>
              ) : null}
            </div>
            <div className="passport-card">
              <div><span>Последняя строка</span><strong>{selectedSample.Timestamp || "—"}</strong></div>
              <div><span>Повторений в файле</span><strong>{selectedSample.SourceOccurrences ?? 0}</strong></div>
              <div><span>Kгаз другой установки</span><strong>{formatNumber(selectedSample.KGasMD)}</strong></div>
              <div><span>Kкл другой установки</span><strong>{formatNumber(selectedSample.KKlinkenbergMD)}</strong></div>
              <div><span>b другой установки, psi</span><strong>{formatNumber(selectedSample.BPsi)}</strong></div>
              <div><span>Объем пор, см3</span><strong>{formatNumber(selectedSample.PoreVolumeCM3)}</strong></div>
            </div>
          </div>
        </section>

        <section className="right-column">
          <div className="panel metrics-panel">
            <div className="metric">
              <span>Display K, mD</span>
              <strong>{formatNumber(calculation.DisplayKMD)}</strong>
            </div>
            <div className="metric">
              <span>Display b, psi</span>
              <strong>{formatNumber(calculation.DisplayBPsi)}</strong>
            </div>
            <div className="metric">
              <span>Recommendation</span>
              <strong>{methodLabel(calculation.DisplayMethod)}</strong>
            </div>
            <div className="metric">
              <span>Processed rows</span>
              <strong>{calculation.ProcessedRows ?? 0}</strong>
            </div>
            <div className="metric">
              <span>K source / b source</span>
              <strong>{`${sourceLabel(calculation.Assessment?.KSource)} / ${sourceLabel(calculation.Assessment?.BSource)}`}</strong>
            </div>
          </div>

          <div className="tabs panel">
            <button className={view === "results" ? "active" : ""} onClick={() => setView("results")}>Результаты</button>
            <button className={view === "logs" ? "active" : ""} onClick={() => setView("logs")}>Логи</button>
          </div>

          {view === "results" ? (
            <div className="results-grid">
              <div className="panel chart-panel">
                <div className="section-title">Кривая падения давления после preprocessing</div>
                <div className="chart-host" ref={processedHostRef}>
                  <ReactECharts option={processedChartOptions} ref={processedChartRef} style={{ height: "100%", width: "100%" }} />
                </div>
              </div>
              <div className="panel chart-panel">
                <div className="section-title">B-22 linearity</div>
                <div className="chart-host" ref={b22HostRef}>
                  <ReactECharts option={b22ChartOptions} ref={b22ChartRef} style={{ height: "100%", width: "100%" }} />
                </div>
              </div>
              <div className="panel details-panel">
                <div className="section-title">Ветви расчета</div>
                <div className="details-grid">
                  <div><span>Full K</span><strong>{formatNumber(calculation.Full?.KMD)}</strong></div>
                  <div><span>Full b, psi</span><strong>{formatNumber(calculation.Full?.BPsi)}</strong></div>
                  <div><span>Full R²</span><strong>{formatNumber(calculation.Full?.R2, 5)}</strong></div>
                  <div><span>Full SE</span><strong>{formatNumber(calculation.Full?.SE, 5)}</strong></div>
                  <div><span>B-22 K</span><strong>{formatNumber(calculation.B22?.KMD)}</strong></div>
                  <div><span>B-22 b, psi</span><strong>{formatNumber(calculation.B22?.BPsi)}</strong></div>
                  <div><span>B-22 R²</span><strong>{formatNumber(calculation.B22?.R2, 5)}</strong></div>
                  <div><span>Max residual</span><strong>{formatNumber(calculation.B22?.MaxResidual, 6)}</strong></div>
                  <div><span>Relative K gap</span><strong>{formatNumber(calculation.Assessment?.RelativeKGap, 5)}</strong></div>
                  <div><span>Mean Pm, psi</span><strong>{formatNumber(calculation.Assessment?.MeanPmPsi, 4)}</strong></div>
                </div>
              </div>
              <div className="panel warnings-panel">
                <div className="section-title">Оценка и пояснения</div>
                <div className="pill-row">
                  <span className={calculation.Assessment?.B22Valid ? "pill ok" : "pill"}>B-22 надёжен</span>
                  <span className={calculation.Assessment?.B22Marginal ? "pill warn" : "pill"}>B-22 пограничный</span>
                  <span className={calculation.Assessment?.FullBCollapsed ? "pill warn" : "pill"}>b во full-fit схлопнулся</span>
                  <span className={calculation.Assessment?.HighPerm ? "pill ok" : "pill"}>Высокая проницаемость</span>
                </div>
                <ul className="notes-list">
                  {(calculation.Assessment?.Rationale ?? []).map((item) => <li key={item}>{item}</li>)}
                  {(calculation.DisplayWarnings ?? []).map((item) => <li key={item}>{item}</li>)}
                </ul>
              </div>
            </div>
          ) : (
            <div className="panel logs-panel">
              <div className="section-title">Backend события</div>
              <div className="logs-list">
                {logs.map((log, index) => (
                  <div className={`log-line level-${(log.Level ?? "INFO").toLowerCase()}`} key={`${log.Time}-${index}`}>
                    <span>{log.Time}</span>
                    <strong>{log.Level}</strong>
                    <span>{log.Message}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </section>
      </main>
    </div>
  );
}

export default App;
