export type Position = { x: number; y: number };
import type { MapTile } from "@/components/ProceduralMap/generator";

export type Cell = MapTile;
export type Effect = {
  id: string;
  name: string;
  description: string;
  negative: boolean;
};
export type Weapon = {
  id: string;
  name: string; icon: string; damage: number; range: number;
  lifePenalty?: number; staminaPenalty: number; toolPenalty: number; description: string;
  effect: Effect[];
};
export type Ability = {
  id: string; name: string; icon: string;
  uses: "single" | "unlimited";
  range: number; damage: number; shield: number; areaSize: number;
  maxActivePerMatch: number; maxUsesPerMatch: number;
  movement?: { forwardDistance: number; climbDistance: number; climbLevels: number };
  description: string; effect: Effect[];
};
export type Unit = {
  id: string; lp: number; icon: string; shield: number; stamina: number;
  weapons: Weapon[]; tools: string[];
  itemSlots: number; abilitySlots: number; abilities: Ability[];
};
export type UnitInstance = Position & {
  player: number; ownerID: string;
  path?: Position[];
  currentStamina: number; currentShield: number; currentLife: number; unit: Unit;
  hasAttacked: boolean;
};
export type Player = { name: string; accoID: string; connected: boolean; units: UnitInstance[] };
export type GameStatus = "waiting" | "playing" | "paused" | "finished";
export type GameStatePayload = {
  id: string; players: Record<string, Player>; board: Cell[];
  turnOrder: string[]; currentTurn: string | null; turnCounter: number;
  status: GameStatus; winner: string | null; revision: number;
};
