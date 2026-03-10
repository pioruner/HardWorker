import { create } from "zustand";

export type CursorMode = "start" | "reper" | "front";

export type AkipSnapshot = {
  connected: boolean;
  lastResponse: string;
  address: string;
  timeBase: number;
  hOffset: string;
  reper: string;
  square: string;
  minY: string;
  minMove: string;
  autoSearch: boolean;
  cursorMode: CursorMode;
  cursorPos: [number, number, number];
  x: number[];
  y: number[];
  vSpeed: string;
  vTime: string;
  volume: string;
  registration: boolean;
};

export type LogEntry = {
  time: string;
  level: "INFO" | "WARN" | "ERROR" | "DEBUG";
  message: string;
};

export type AkipControls = {
  address: string;
  timeBase: number;
  hOffset: string;
  reper: string;
  square: string;
  minY: string;
  minMove: string;
  autoSearch: boolean;
  cursorMode: CursorMode;
  cursorPos: [number, number, number];
};

type AkipState = {
  snapshot: AkipSnapshot;
  logs: LogEntry[];
  setSnapshot: (snapshot: AkipSnapshot) => void;
  setLogs: (logs: LogEntry[]) => void;
  patchControls: (next: Partial<AkipControls>) => void;
};

export const defaultSnapshot: AkipSnapshot = {
  connected: false,
  lastResponse: "Ожидание подключения к прибору...",
  address: "192.168.0.100:3000",
  timeBase: 3,
  hOffset: "0",
  reper: "25",
  square: "10",
  minY: "20",
  minMove: "0.3",
  autoSearch: false,
  registration: false,
  cursorMode: "start",
  cursorPos: [18, 34, 62],
  x: [0, 1, 2, 3],
  y: [1, 1, 1, 1],
  vSpeed: "0.00",
  vTime: "0.00",
  volume: "0.00",
};

export const useAkipStore = create<AkipState>((set) => ({
  snapshot: defaultSnapshot,
  logs: [],
  setSnapshot: (snapshot) => set({ snapshot }),
  setLogs: (logs) => set({ logs }),
  patchControls: (next) =>
    set((state) => ({
      snapshot: {
        ...state.snapshot,
        ...next,
      },
    })),
}));

export function toControls(snapshot: AkipSnapshot): AkipControls {
  return {
    address: snapshot.address,
    timeBase: snapshot.timeBase,
    hOffset: snapshot.hOffset,
    reper: snapshot.reper,
    square: snapshot.square,
    minY: snapshot.minY,
    minMove: snapshot.minMove,
    autoSearch: snapshot.autoSearch,
    cursorMode: snapshot.cursorMode,
    cursorPos: snapshot.cursorPos,
  };
}
