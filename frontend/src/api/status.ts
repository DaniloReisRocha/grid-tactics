export const GameStatus = {
  waiting: "waiting",
  playing: "playing",
  paused: "paused",
  finished: "finished",
} as const;

export type GameStatus = (typeof GameStatus)[keyof typeof GameStatus];

export const ConnectionStatus = {
  disconnected: "disconnected",
  connecting: "connecting",
  connected: "connected",
  reconnecting: "reconnecting",
} as const;

export type ConnectionStatus = (typeof ConnectionStatus)[keyof typeof ConnectionStatus];
