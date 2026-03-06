import { useMemo } from "react";
import ReactECharts from "echarts-for-react";
import "./App.scss";
import { calcWaveMetrics, getCurrentCursorValue, useAkipStore } from "./store/akipStore";

const timeScaleValues = ["1us", "2us", "5us", "10us", "20us", "50us", "100us"];

function App() {
  const {
    connected,
    timeBase,
    hOffset,
    reper,
    square,
    minY,
    minMove,
    autoSearch,
    registration,
    cursorMode,
    cursorPos,
    setField,
    setTimeBase,
    setCursorMode,
    setCurrentCursorPos,
    toggleAutoSearch,
    toggleRegistration,
  } = useAkipStore();

  const currentCursorPos = getCurrentCursorValue({ cursorMode, cursorPos });
  const { waveTime, speed, volume } = calcWaveMetrics({ cursorPos, reper, square });

  const seriesData = useMemo(() => {
    const data: number[][] = [];
    for (let i = 0; i <= 100; i += 0.2) {
      const noise = Math.sin(i * 1.4) * 2.6;
      const wave = 62 * Math.sin((i - 14) * 0.11) * Math.exp(-Math.pow((i - 55) / 26, 2));
      const peak = 34 * Math.exp(-Math.pow((i - 63) / 1.9, 2));
      data.push([Number(i.toFixed(2)), Number((noise + wave + peak).toFixed(2))]);
    }
    return data;
  }, []);

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
        min: 0,
        max: 100,
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
          data: seriesData,
          markLine: {
            symbol: "none",
            label: { show: true, color: "#ecf2fa", fontSize: 11 },
            data: [
              {
                name: "Начало",
                xAxis: cursorPos.start,
                lineStyle: { color: "#7cc576", width: 2 },
              },
              {
                name: "Репер",
                xAxis: cursorPos.reper,
                lineStyle: { color: "#ffc857", width: 2 },
              },
              {
                name: "Граница",
                xAxis: cursorPos.front,
                lineStyle: { color: "#ff6a6a", width: 2 },
              },
            ],
          },
        },
      ],
    }),
    [cursorPos, seriesData],
  );

  return (
    <main className="akip-layout">
      <header className="panel panel-top">
        <label className="field">
          <span>Развертка</span>
          <select value={timeBase} onChange={(e) => setTimeBase(e.target.value)}>
            {timeScaleValues.map((value) => (
              <option key={value} value={value}>
                {value}
              </option>
            ))}
          </select>
        </label>

        <label className="field">
          <span>Смещение (мкс)</span>
          <input value={hOffset} onChange={(e) => setField("hOffset", e.target.value)} />
        </label>

        <label className="field">
          <span>Репер dL (см)</span>
          <input value={reper} onChange={(e) => setField("reper", e.target.value)} />
        </label>

        <label className="field">
          <span>Площадь трубки (см²)</span>
          <input value={square} onChange={(e) => setField("square", e.target.value)} />
        </label>

        <div className={`status ${connected ? "is-online" : "is-offline"}`}>
          {connected ? "Связь: онлайн" : "Связь: оффлайн"}
        </div>
      </header>

      <section className="panel panel-metrics">
        <div className="metric">
          <span>Скорость волны</span>
          <strong>{speed.toFixed(2)} м/с</strong>
        </div>
        <div className="metric">
          <span>Время волны</span>
          <strong>{waveTime.toFixed(2)} мкс</strong>
        </div>
        <div className="metric">
          <span>Объём фазы</span>
          <strong>{volume.toFixed(2)} см³</strong>
        </div>
        <button className={`toggle ${registration ? "is-active" : ""}`} onClick={toggleRegistration}>
          {registration ? "Регистрация: вкл" : "Регистрация: выкл"}
        </button>
      </section>

      <section className="panel panel-chart">
        <div className="panel-title">Осциллограмма</div>
        <div className="chart-host">
          <ReactECharts option={chartOptions} style={{ height: "100%", width: "100%" }} />
        </div>
      </section>

      <section className="panel panel-bottom">
        <label className="slider-wrap">
          <span>Позиция курсора ({cursorMode})</span>
          <input
            type="range"
            min={0}
            max={100}
            step={0.1}
            value={currentCursorPos}
            onChange={(e) => setCurrentCursorPos(Number(e.target.value))}
          />
          <b>{currentCursorPos.toFixed(2)} мкс</b>
        </label>

        <div className="cursor-buttons">
          <button className={cursorMode === "start" ? "is-active" : ""} onClick={() => setCursorMode("start")}>
            Курсор: Старт
          </button>
          <button className={cursorMode === "reper" ? "is-active" : ""} onClick={() => setCursorMode("reper")}>
            Курсор: Репер
          </button>
          <button className={cursorMode === "front" ? "is-active" : ""} onClick={() => setCursorMode("front")}>
            Курсор: Граница
          </button>
        </div>

        <label className="field inline">
          <span>Мин. Y</span>
          <input value={minY} onChange={(e) => setField("minY", e.target.value)} />
        </label>
        <label className="field inline">
          <span>Мин. смещение</span>
          <input value={minMove} onChange={(e) => setField("minMove", e.target.value)} />
        </label>

        <button className={`toggle ${autoSearch ? "is-active" : ""}`} onClick={toggleAutoSearch}>
          {autoSearch ? "Автопоиск: вкл" : "Автопоиск: выкл"}
        </button>
      </section>
    </main>
  );
}

export default App;
