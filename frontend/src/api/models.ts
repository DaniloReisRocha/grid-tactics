export type Cell = {
  x: number;
  y: number;
  elevation: 0 | 1 | 2;
};

export type Unit = {
  id: string;
  lp: number;
  icon: string;
  shield: number;
  stamina: number;
  weapons: Weapon[];
  tools: any[];
};

export type UnitInstance = {
  x: number;
  y: number;
  player: number;
  currentStamina: number;
  currentShield: number;
  currentLife: number;
  unit: Unit;
};

export type Weapon = {
  name: string;
  icon: string;
  damage: number;
  range: number;
  staminaPenalty: number;
  toolPenalty: number;
  description: string;
  canAttackAdjacent?: boolean;
  effect?: (attacker: UnitInstance, target: UnitInstance) => UnitInstance;
};

export type Player = {
  name: string;
  accoID: string;
  units: UnitInstance[];
};

export type GameStatePayload = {
  id: string;
  players: Record<string, Player>;
  board: Cell[];
  turnOrder: string[];
  currentTurn: string | null;
  turnCounter: number;
};
