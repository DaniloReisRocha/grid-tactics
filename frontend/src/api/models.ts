export type Position = { x: number; y: number };
import type { MapTile } from "@/components/ProceduralMap/generator";
import type { GameStatus } from "./status";
export type { GameStatus } from "./status";

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
  passive: boolean;
  id: string; name: string; icon: string;
  range: number; damage: number; shield: number; areaSize: number;
  heal: number;
  maxUsesPerMatch: number;
  movement?: { forwardDistance: number; climbDistance: number; climbLevels: number };
  description: string; effect: Effect[];
};
export type Unit = {
  id: string; lp: number; icon: string; shield: number; stamina: number;
  weapons: Weapon[]; tools: string[];
  itemSlots: number; abilitySlots: number; abilities: Ability[];
};
export type UnitInstance = Position & {
  turnCount: number;
  player: number; ownerID: string;
  path?: Position[];
  currentStamina: number; currentShield: number; currentLife: number; unit: Unit;
  hasAttacked: boolean;
  abilityUses?: Record<string,number>;
  movementMode?: "jump" | "";
};
export type Player = { name: string; accoID: string; connected: boolean; units: UnitInstance[] };
export type GameStatePayload = {
  serverTimeMS: number; turnDeadlineMS: number; turnRemainingMS: number;
  orbs: (Position & { id: string; triggered: boolean; collectedBy?: string })[];
  currentUnitID: string; roundCounter: number; roundTurn: number;
  roundOrder: { playerID: string; unitID: string }[];
  id: string; players: Record<string, Player>; board: Cell[];
  turnOrder: string[]; currentTurn: string | null; turnCounter: number;
  status: GameStatus; winner: string | null; revision: number;
  matchAbilityUses: Record<string,number>;
};
