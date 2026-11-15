import { create } from "zustand";
import type { GameStatePayload } from "@/api/models";
import type { UnitLoadout } from "@/api/team";
import { GameSocket, MessageType, type ConnectionStatus } from "@/api/webSocket/GameSocket";
import { canUnitAttack, isAttackTarget, isBarricadeTarget } from "@/internal/game/combat";

const emptyGame: GameStatePayload = {
  id: "", players: {}, board: [], turnOrder: [], currentTurn: null,
  turnCounter: 0, status: "waiting", winner: null, revision: -1,
  matchAbilityUses: {},
};
type GameStore = GameStatePayload & {
  accountID: string; unitID: string | null; attackMode: boolean;
  skillMode: string | null;
  setSkillMode: (id: string | null) => void;
  activateAbility: (unitID: string, abilityID: string, x: number, y: number) => void;
  connection: ConnectionStatus; error: string | null; pending: boolean;
  setUnit: (unitID: string) => void;
  setAttackMode: (active: boolean) => void;
  joinGame: (sessionID: string, accoID: string, team?: UnitLoadout[]) => void;
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
    ...emptyGame, accountID: "", unitID: null, attackMode: false, skillMode: null,
    connection: "disconnected", error: null, pending: false,
    setUnit: (unitID) => set({ unitID, attackMode: false, skillMode: null }),
    setAttackMode: (attackMode) => set({ attackMode, skillMode: null }),
    setSkillMode: (skillMode) => set({ skillMode, attackMode: false }),
    activateAbility: (unitID,abilityID,x,y) => act(MessageType.useAbility,{unitID,abilityID,to:{x,y}}),
    clearError: () => set({ error: null }),
    joinGame: (sessionID, accountID, team) => {
      set({ ...emptyGame, id: sessionID, accountID, unitID: null, attackMode: false, skillMode:null, error: null, pending: false });
      GameSocket.connect(sessionID, accountID, team);
    },
    leaveGame: (forfeit = false) => {
      if (forfeit) GameSocket.sendAction(MessageType.leaveGame);
      GameSocket.disconnect();
      set({ ...emptyGame, accountID: "", unitID: null, attackMode: false, skillMode:null, error: null, pending: false });
    },
    movePiece: (unitID, x, y) => act(MessageType.movePiece, { unitID, to: { x, y } }),
    attackPiece: (unitID, x, y) => {
      const state = get();
      const attacker = state.players[state.accountID]?.units.find((unit) => unit.unit.id === unitID);
      const target = Object.values(state.players).flatMap((player) => player.units)
        .find((unit) => unit.x === x && unit.y === y && unit.currentLife > 0);
      // An empty/invalid slot must not send an action or change attack mode/pending state.
      const tile=state.board.find((cell)=>cell.x===x&&cell.y===y);
      if (!attacker || (!isAttackTarget(attacker,target,state.board) && !(tile&&isBarricadeTarget(attacker,tile,state.board)))) return;
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
      skillMode: null,
      board: JSON.stringify(payload.board) === JSON.stringify(state.board) ? state.board : payload.board,
      unitID: selected ? state.unitID : null,
      attackMode: state.attackMode && payload.currentTurn === state.accountID && payload.status === "playing"
        && Boolean(selected && canUnitAttack(selected)),
    });
  }
});
export const useGameSocket = gameStore;
