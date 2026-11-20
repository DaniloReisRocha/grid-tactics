import { GameStatus, ConnectionStatus } from "@/api/status";
import { create } from "zustand";
import type { GameStatePayload } from "@/api/models";
import type { UnitLoadout } from "@/api/team";
import { GameSocket, MessageType } from "@/api/webSocket/GameSocket";
import { canUnitAttack, isAttackTarget, isBarricadeTarget } from "@/internal/game/combat";
import type { MapExplosion } from "@/components/ProceduralMap/types";

const emptyGame: GameStatePayload = {
  grenadeEvent: null,
  serverTimeMS: 0, turnDeadlineMS: 0, turnRemainingMS: 0,
  orbs: [],
  currentUnitID: "", roundCounter: 0, roundTurn: 0, roundOrder: [],
  id: "", players: {}, board: [], turnOrder: [], currentTurn: null,
  turnCounter: 0, status: GameStatus.waiting, winner: null, revision: -1,
  matchAbilityUses: {},
};
type GameStore = GameStatePayload & {
  grenadeExplosions: MapExplosion[];
  finishGrenadeExplosion: (id: string) => void;
  serverClockOffsetMS: number;
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
    if (state.pending || state.status !== GameStatus.playing || state.currentTurn !== state.accountID) return;
    if (type !== MessageType.endTurn && (payload as { unitID?: string }).unitID !== state.currentUnitID) return;
    if (!GameSocket.sendAction(type, payload)) { set({ error: "Aguarde a conexão com o servidor." }); return; }
    set({ pending: true, error: null });
  };
  return {
    ...emptyGame, accountID: "", unitID: null, attackMode: false, skillMode: null,
    serverClockOffsetMS: 0,
    grenadeExplosions: [],
    finishGrenadeExplosion: (id) => set((state) => ({grenadeExplosions: state.grenadeExplosions.filter((item) => item.id !== id)})),
    connection: ConnectionStatus.disconnected, error: null, pending: false,
    setUnit: (unitID) => set({ unitID, attackMode: false, skillMode: null }),
    setAttackMode: (attackMode) => set({ attackMode, skillMode: null }),
    setSkillMode: (skillMode) => set({ skillMode, attackMode: false }),
    activateAbility: (unitID,abilityID,x,y) => act(MessageType.useAbility,{unitID,abilityID,to:{x,y}}),
    clearError: () => set({ error: null }),
    joinGame: (sessionID, accountID, team) => {
      set({ ...emptyGame, grenadeExplosions: [], id: sessionID, accountID, unitID: null, attackMode: false, skillMode:null, error: null, pending: false });
      GameSocket.connect(sessionID, accountID, team);
    },
    leaveGame: (forfeit = false) => {
      if (forfeit) GameSocket.sendAction(MessageType.leaveGame);
      GameSocket.disconnect();
      set({ ...emptyGame, grenadeExplosions: [], accountID: "", unitID: null, attackMode: false, skillMode:null, error: null, pending: false });
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
    endTurn: () => act(MessageType.endTurn, {turnCounter: get().turnCounter}),
  };
});

GameSocket.onStatus((connection) => gameStore.setState({ connection, ...(connection !== ConnectionStatus.connected ? { pending: false } : {}) }));
GameSocket.subscribe((message) => {
  const state = gameStore.getState();
  if (message.type === "actionFailed") {
    gameStore.setState({ error: message.payload.message, pending: false });
  } else if (message.type === "gameStateUpdate") {
    const payload = message.payload;
    if (payload.id !== state.id || payload.revision < state.revision) return;
    const selected = Object.values(payload.players).flatMap((player) => player.units)
      .find((unit) => unit.unit.id === state.unitID && unit.currentLife > 0);
    const explosion = payload.grenadeEvent;
    const newExplosion = explosion && state.revision >= 0 && explosion.id !== state.grenadeEvent?.id
      && payload.serverTimeMS - explosion.createdAtMS <= 5000;
    gameStore.setState({
      ...payload, pending: false, error: null,
      grenadeExplosions: newExplosion ? [...state.grenadeExplosions, {...explosion, id: String(explosion.id)}] : state.grenadeExplosions,
      serverClockOffsetMS: payload.serverTimeMS ? payload.serverTimeMS - Date.now() : 0,
      skillMode: null,
      board: JSON.stringify(payload.board) === JSON.stringify(state.board) ? state.board : payload.board,
      unitID: payload.currentUnitID !== state.currentUnitID ? payload.currentUnitID || null : selected ? state.unitID : null,
      attackMode: state.attackMode && payload.currentTurn === state.accountID && payload.status === GameStatus.playing
        && payload.currentUnitID === state.currentUnitID && selected?.unit.id === payload.currentUnitID
        && Boolean(selected && canUnitAttack(selected)),
    });
  }
});
export const useGameSocket = gameStore;
