export type Position = { x: number; y: number };
export type Cell = Position & { elevation: 0 | 1 | 2 };
export type Weapon = {
  name: string; icon: string; damage: number; range: number;
  lifePenalty?: number; staminaPenalty: number; toolPenalty: number; description: string;
};
export type Unit = {
  id: string; lp: number; icon: string; shield: number; stamina: number;
  weapons: Weapon[]; tools: string[];
};
export type UnitInstance = Position & {
  player: number; ownerID: string;
  currentStamina: number; currentShield: number; currentLife: number; unit: Unit;
};
export type Player = { name: string; accoID: string; connected: boolean; units: UnitInstance[] };
export type GameStatus = "waiting" | "playing" | "paused" | "finished";
export type GameStatePayload = {
  id: string; players: Record<string, Player>; board: Cell[];
  turnOrder: string[]; currentTurn: string | null; turnCounter: number;
  status: GameStatus; winner: string | null; revision: number;
};
