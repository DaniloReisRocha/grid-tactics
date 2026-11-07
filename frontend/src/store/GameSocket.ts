import type { Cell, GameStatePayload, Player } from "@/api/models";
import { create, useStore } from "zustand";
import { immer } from "zustand/middleware/immer";

const GameSocket = new WebSocket("ws://localhost:8080/api/game-connect");

const socketSend = (messageType: string, payload: any) => {
  GameSocket.send(JSON.stringify({ type: messageType, payload: payload }));
};

const MessageType = {
  echo: "echo",
  joinGame: "joinGame",
  movePiece: "movePiece",
  actionFailed: "actionFailed",
  gameStateUpdate: "gameStateUpdate",
} as const;

type Message = {
  type: keyof typeof MessageType;
  payload: any;
};

type Store = {
  unitID?: string | null;
  gameID: string;
  players: Record<string, Player | undefined>;
  board: Cell[];
  turnOrder: string[];
  currentTurn: string | null;
  turnCounter: number;
};

type Actions = {
  setUnit: (unitID: string) => void;
  joinGame: (sessionID: string, accoID: string) => void;
  movePiece: (unitID: string, x: number, y: number) => void;
  setGameState: (payload: GameStatePayload) => void;
  reset: () => void;
};

const initialValues: Store = {
  unitID: null,
  gameID: "",
  players: {},
  board: [],
  turnOrder: [],
  currentTurn: null,
  turnCounter: 0,
};

const gameStore = create<Store & Actions>()(
  immer((set) => ({
    ...initialValues,
    setUnit: (unitID) =>
      set((state) => {
        state.unitID = unitID;
      }),
    setGameState: (payload) =>
      set((state) => {
        state.gameID = payload.id;
        state.players = payload.players;
        state.board = payload.board;
        state.turnOrder = payload.turnOrder;
        state.currentTurn = payload.currentTurn;
        state.turnCounter = payload.turnCounter;
      }),
    reset: () =>
      set((state) => {
        state.gameID = initialValues.gameID;
        state.players = initialValues.players;
        state.board = initialValues.board;
        state.turnOrder = initialValues.turnOrder;
        state.currentTurn = initialValues.currentTurn;
        state.turnCounter = initialValues.turnCounter;
      }),
    joinGame: (sessionID, accoID) => {
      socketSend(MessageType.joinGame, {
        sessionID,
        accoID,
      });
    },
    movePiece: (unitID, x, y) => {
      socketSend(MessageType.movePiece, {
        unitID,
        to: { x, y },
      });
    },
  }))
);

GameSocket.onmessage = (event) => {
  const { type, payload }: Message = JSON.parse(event.data);
  const { setGameState } = gameStore.getState();

  switch (type) {
    case MessageType.gameStateUpdate:
      setGameState(payload);
      break;
    case MessageType.actionFailed:
      console.error(payload.message);
      break;
  }
};

export const useGameSocket = () => useStore(gameStore, (s) => s);
