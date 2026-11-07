import { create } from "zustand";
import type { GameStatePayload } from "@/api/models";
import { GameSocket, MessageType, type ConnectionStatus } from "@/api/webSocket/GameSocket";
import { canUnitAttack, isAttackTarget } from "@/internal/game/combat";

const emptyGame: GameStatePayload = {
  id: "", players: {}, board: [], turnOrder: [], currentTurn: null,
  turnCounter: 0, status: "waiting", winner: null, revision: -1,
};
type GameStore = GameStatePayload & {
  accountID: string; unitID: string | null; attackMode: boolean;
  connection: ConnectionStatus; error: string | null; pending: boolean;
  setUnit: (unitID: string) => void;
  setAttackMode: (active: boolean) => void;
  joinGame: (sessionID: string, accoID: string) => void;
  leaveGame: (forfeit?: boolean) => void;
  movePiece: (unitID: string, x: number, y: number) => void;
  attackPiece: (unitID: string, x: number, y: number) => void;
  endTurn: () => void;
  clearError: () => void;
};

export const gameStore = create<GameStore>((set, get) => {
  const act = (type: string, payload: unknown = {}) => {
    const state = get();
    if (state.pending || state.status !== "playing" || state.currentTurn !== state.accountID) return;
    if (!GameSocket.sendAction(type, payload)) { set({ error: "Aguarde a conexão com o servidor." }); return; }
    set({ pending: true, error: null });
  };
  return {
    ...emptyGame, accountID: "", unitID: null, attackMode: false,
    connection: "disconnected", error: null, pending: false,
    setUnit: (unitID) => set({ unitID, attackMode: false }),
    setAttackMode: (attackMode) => set({ attackMode }),
    clearError: () => set({ error: null }),
    joinGame: (sessionID, accountID) => {
      set({ ...emptyGame, id: sessionID, accountID, unitID: null, attackMode: false, error: null, pending: false });
      GameSocket.connect(sessionID, accountID);
    },
    leaveGame: (forfeit = false) => {
      if (forfeit) GameSocket.sendAction(MessageType.leaveGame);
      GameSocket.disconnect();
      set({ ...emptyGame, accountID: "", unitID: null, attackMode: false, error: null, pending: false });
    },
    movePiece: (unitID, x, y) => act(MessageType.movePiece, { unitID, to: { x, y } }),
    attackPiece: (unitID, x, y) => {
      const state = get();
      const attacker = state.players[state.accountID]?.units.find((unit) => unit.unit.id === unitID);
      const target = Object.values(state.players).flatMap((player) => player.units)
        .find((unit) => unit.x === x && unit.y === y && unit.currentLife > 0);
      // An empty/invalid slot must not send an action or change attack mode/pending state.
      if (!attacker || !isAttackTarget(attacker, target)) return;
      act(MessageType.attackPiece, { unitID, to: { x, y } });
    },
    endTurn: () => act(MessageType.endTurn),
  };
});

GameSocket.onStatus((connection) => gameStore.setState({ connection, ...(connection !== "connected" ? { pending: false } : {}) }));
GameSocket.subscribe((message) => {
  const state = gameStore.getState();
  if (message.type === "actionFailed") {
    gameStore.setState({ error: message.payload.message, pending: false });
  } else if (message.type === "gameStateUpdate") {
    const payload = message.payload;
    if (payload.id !== state.id || payload.revision < state.revision) return;
    const selected = Object.values(payload.players).flatMap((player) => player.units)
      .find((unit) => unit.unit.id === state.unitID && unit.currentLife > 0);
    gameStore.setState({
      ...payload, pending: false, error: null,
      board: JSON.stringify(payload.board) === JSON.stringify(state.board) ? state.board : payload.board,
      unitID: selected ? state.unitID : null,
      attackMode: state.attackMode && payload.currentTurn === state.accountID && payload.status === "playing"
        && Boolean(selected && canUnitAttack(selected)),
    });
  }
});
export const useGameSocket = gameStore;
