import { create } from "zustand";

export type CursorMode = "start" | "reper" | "front";

type AkipState = {
  connected: boolean;
  timeBase: string;
  hOffset: string;
  reper: string;
  square: string;
  minY: string;
  minMove: string;
  autoSearch: boolean;
  registration: boolean;
  cursorMode: CursorMode;
  cursorPos: Record<CursorMode, number>;
  setField: (field: "hOffset" | "reper" | "square" | "minY" | "minMove", value: string) => void;
  setTimeBase: (value: string) => void;
  setCursorMode: (value: CursorMode) => void;
  setCurrentCursorPos: (value: number) => void;
  setCursorPos: (mode: CursorMode, value: number) => void;
  toggleAutoSearch: () => void;
  toggleRegistration: () => void;
};

export const useAkipStore = create<AkipState>((set, get) => ({
  connected: true,
  timeBase: "10us",
  hOffset: "0",
  reper: "25",
  square: "10",
  minY: "20",
  minMove: "0.3",
  autoSearch: true,
  registration: false,
  cursorMode: "start",
  cursorPos: {
    start: 18,
    reper: 34,
    front: 62,
  },
  setField: (field, value) => set({ [field]: value }),
  setTimeBase: (value) => set({ timeBase: value }),
  setCursorMode: (value) => set({ cursorMode: value }),
  setCurrentCursorPos: (value) =>
    set((state) => ({
      cursorPos: {
        ...state.cursorPos,
        [state.cursorMode]: value,
      },
    })),
  setCursorPos: (mode, value) =>
    set((state) => ({
      cursorPos: {
        ...state.cursorPos,
        [mode]: value,
      },
    })),
  toggleAutoSearch: () => set((state) => ({ autoSearch: !state.autoSearch })),
  toggleRegistration: () => set((state) => ({ registration: !state.registration })),
}));

export function calcWaveMetrics(state: Pick<AkipState, "cursorPos" | "reper" | "square">) {
  const waveTime = Math.max(0, state.cursorPos.front - state.cursorPos.start);
  const reperDelta = Math.max(0.001, state.cursorPos.reper - state.cursorPos.start);
  const reperValue = Number.parseFloat(state.reper) || 0;
  const squareValue = Number.parseFloat(state.square) || 0;
  const speed = (reperValue / reperDelta) * 10;
  const volume = (waveTime * reperValue * squareValue) / 1000;

  return {
    waveTime,
    speed,
    volume,
  };
}

export function getCurrentCursorValue(state: Pick<AkipState, "cursorMode" | "cursorPos">) {
  return state.cursorPos[state.cursorMode];
}
