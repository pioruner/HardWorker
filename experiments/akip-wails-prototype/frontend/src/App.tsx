import { useEffect, useMemo } from "react";
import ReactECharts from "echarts-for-react";
import "./App.scss";
import { ApplyControls, GetSnapshot } from "../wailsjs/go/main/App";
import { defaultSnapshot, toControls, useAkipStore, type AkipSnapshot, type CursorMode } from "./store/akipStore";

const timeScaleValues = ["1us", "2us", "5us", "10us", "20us", "50us", "100us"];
const cursorModes: CursorMode[] = ["start", "reper", "front"];

const cursorIndexByMode: Record<CursorMode, number> = {
  start: 0,
  reper: 1,
  front: 2,
};

function asSnapshot(value: unknown): AkipSnapshot {
  if (!value || typeof value !== "object") {
    return defaultSnapshot;
  }
  const raw = value as Partial<AkipSnapshot>;
  return { ...defaultSnapshot, ...raw };
}

function App() {
  const snapshot = useAkipStore((state) => state.snapshot);
  const setSnapshot = useAkipStore((state) => state.setSnapshot);
  const patchControls = useAkipStore((state) => state.patchControls);

  useEffect(() => {
    let active = true;

    const sync = async () => {
      try {
        const next = asSnapshot(await GetSnapshot());
        if (active) {
          setSnapshot(next);
        }
      } catch {
        // ignore polling errors in UI loop
      }
    };

    void sync();
    const id = window.setInterval(sync, 250);
    return () => {
      active = false;
      window.clearInterval(id);
    };
  }, [setSnapshot]);

  const applyCurrentControls = async () => {
    const current = useAkipStore.getState().snapshot;
    try {
      const next = asSnapshot(await ApplyControls(toControls(current)));
      setSnapshot(next);
    } catch {
      // backend may be restarting while reconnecting
    }
  };

  const patchAndApply = (next: Partial<AkipSnapshot>) => {
    patchControls(next);
    void applyCurrentControls();
  };

  const activeCursorIndex = cursorIndexByMode[snapshot.cursorMode];
  const currentCursorPos = snapshot.cursorPos[activeCursorIndex];

  const chartSeriesData = useMemo(() => {
    const points = Math.min(snapshot.x.length, snapshot.y.length);
    if (points <= 0) {
      return [[0, 0]];
    }
    const data = new Array(points);
    for (let i = 0; i < points; i += 1) {
      data[i] = [snapshot.x[i], snapshot.y[i]];
    }
    return data;
  }, [snapshot.x, snapshot.y]);

  const xMin = snapshot.x.length > 0 ? snapshot.x[0] : 0;
  const xMax = snapshot.x.length > 1 ? snapshot.x[snapshot.x.length - 1] : 100;

  const chartOptions = useMemo(
    () => ({
      animation: false,
      backgroundColor: "transparent",
      grid: {
        top: 20,
        right: 18,
        bottom: 38,
        left: 48,
      },
      tooltip: { trigger: "axis" },
      xAxis: {
        type: "value",
        min: xMin,
        max: xMax,
        axisLabel: { color: "#bdc8d8" },
        splitLine: { lineStyle: { color: "rgba(136, 156, 180, 0.18)" } },
      },
      yAxis: {
        type: "value",
        min: -150,
        max: 150,
        axisLabel: { color: "#bdc8d8" },
        splitLine: { lineStyle: { color: "rgba(136, 156, 180, 0.18)" } },
      },
      series: [
        {
          name: "УЗ Волна",
          type: "line",
          symbol: "none",
          lineStyle: { width: 2, color: "#42d4ff" },
          data: chartSeriesData,
          markLine: {
            symbol: "none",
            label: { show: true, color: "#ecf2fa", fontSize: 11 },
            data: [
              {
                name: "Начало",
                xAxis: snapshot.cursorPos[0],
                lineStyle: { color: "#7cc576", width: 2 },
              },
              {
                name: "Репер",
                xAxis: snapshot.cursorPos[1],
                lineStyle: { color: "#ffc857", width: 2 },
              },
              {
                name: "Граница",
                xAxis: snapshot.cursorPos[2],
                lineStyle: { color: "#ff6a6a", width: 2 },
              },
            ],
          },
        },
      ],
    }),
    [chartSeriesData, snapshot.cursorPos, xMax, xMin],
  );

  return (
    <main className="akip-layout">
      <header className="panel panel-top">
        <label className="field">
          <span>Адрес прибора</span>
          <input value={snapshot.address} onChange={(e) => patchAndApply({ address: e.target.value })} />
        </label>
        <label className="field">
          <span>Развертка</span>
          <select
            value={snapshot.timeBase}
            onChange={(e) => patchAndApply({ timeBase: Number(e.target.value) })}
          >
            {timeScaleValues.map((value, index) => (
              <option key={value} value={index}>
                {value}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          <span>Смещение (мкс)</span>
          <input value={snapshot.hOffset} onChange={(e) => patchAndApply({ hOffset: e.target.value })} />
        </label>
        <label className="field">
          <span>Репер dL (см)</span>
          <input value={snapshot.reper} onChange={(e) => patchAndApply({ reper: e.target.value })} />
        </label>
        <label className="field">
          <span>Площадь трубки (см²)</span>
          <input value={snapshot.square} onChange={(e) => patchAndApply({ square: e.target.value })} />
        </label>

        <div className={`status ${snapshot.connected ? "is-online" : "is-offline"}`}>
          {snapshot.connected ? "Связь: онлайн" : "Связь: оффлайн"}
        </div>
      </header>

      <section className="panel panel-metrics">
        <div className="metric">
          <span>Скорость волны</span>
          <strong>{snapshot.vSpeed} м/с</strong>
        </div>
        <div className="metric">
          <span>Время волны</span>
          <strong>{snapshot.vTime} мкс</strong>
        </div>
        <div className="metric">
          <span>Объём фазы</span>
          <strong>{snapshot.volume} см³</strong>
        </div>
        <input className="response-box" readOnly value={snapshot.lastResponse} />
      </section>

      <section className="panel panel-chart">
        <div className="panel-title">Осциллограмма</div>
        <div className="chart-host">
          <ReactECharts option={chartOptions} style={{ height: "100%", width: "100%" }} />
        </div>
      </section>

      <section className="panel panel-bottom">
        <label className="slider-wrap">
          <span>Позиция курсора ({snapshot.cursorMode})</span>
          <input
            type="range"
            min={xMin}
            max={xMax}
            step={0.1}
            value={currentCursorPos}
            onChange={(e) => {
              const next = [...snapshot.cursorPos] as [number, number, number];
              next[activeCursorIndex] = Number(e.target.value);
              patchAndApply({ cursorPos: next });
            }}
          />
          <b>{currentCursorPos.toFixed(2)} мкс</b>
        </label>

        <div className="cursor-buttons">
          {cursorModes.map((mode) => (
            <button
              key={mode}
              className={snapshot.cursorMode === mode ? "is-active" : ""}
              onClick={() => patchAndApply({ cursorMode: mode })}
            >
              {mode === "start" ? "Курсор: Старт" : mode === "reper" ? "Курсор: Репер" : "Курсор: Граница"}
            </button>
          ))}
        </div>

        <label className="field inline">
          <span>Мин. Y</span>
          <input value={snapshot.minY} onChange={(e) => patchAndApply({ minY: e.target.value })} />
        </label>
        <label className="field inline">
          <span>Мин. смещение</span>
          <input value={snapshot.minMove} onChange={(e) => patchAndApply({ minMove: e.target.value })} />
        </label>

        <button className={`toggle ${snapshot.autoSearch ? "is-active" : ""}`} onClick={() => patchAndApply({ autoSearch: !snapshot.autoSearch })}>
          {snapshot.autoSearch ? "Автопоиск: вкл" : "Автопоиск: выкл"}
        </button>
      </section>
    </main>
  );
}

export default App;
