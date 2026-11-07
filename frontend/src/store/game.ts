import { create, useStore } from "zustand";
import type { UnitInstance, Cell } from "@/api/models";

type GameState = {
  board: Cell[];
  units: UnitInstance[];
  pieceID: string | null;
  currentPlayer: 1 | 2;
  attackMode: boolean;

  setBoard: (board: Cell[]) => void;
  setPieces: (units: UnitInstance[]) => void;
  setAttackMode: (a: boolean) => void;
  selectPiece: (unit: string | null) => void;
  startTurn: () => void;
  movePiece: (pieceId: string, x: number, y: number) => void;
  attackPiece: (targetX: number, targetY: number) => void;
  endTurn: () => void;
};

type Point = { x: number; y: number };

export function findPathToTarget(
  board: Cell[],
  unit: UnitInstance,
  targetX: number,
  targetY: number
): Point[] | null {
  const open: { x: number; y: number; cost: number; path: Point[] }[] = [
    { x: unit.x, y: unit.y, cost: 0, path: [] },
  ];
  const visited = new Set<string>();

  while (open.length > 0) {
    const current = open.shift()!;
    const key = `${current.x}-${current.y}`;

    if (visited.has(key)) continue;
    visited.add(key);

    if (current.x === targetX && current.y === targetY) {
      return current.path;
    }

    const neighbors = [
      { dx: -1, dy: 0 },
      { dx: 1, dy: 0 },
      { dx: 0, dy: -1 },
      { dx: 0, dy: 1 },
      { dx: -1, dy: -1 },
      { dx: 1, dy: -1 },
      { dx: -1, dy: 1 },
      { dx: 1, dy: 1 },
    ];

    for (const { dx, dy } of neighbors) {
      const nx = current.x + dx;
      const ny = current.y + dy;
      const cell = board.find((c) => c.x === nx && c.y === ny);
      if (!cell) continue;

      const moveCost = 1 + cell.elevation;
      const nextCost = current.cost + moveCost;

      if (nextCost <= unit.currentStamina) {
        open.push({
          x: nx,
          y: ny,
          cost: nextCost,
          path: [...current.path, { x: nx, y: ny }],
        });
      }
    }
  }

  return null;
}

const store = create<GameState>((set, get) => ({
  board: [],
  units: [],
  pieceID: null,
  currentPlayer: 1,
  attackMode: false,

  setBoard: (board) => set({ board }),
  setPieces: (units) => {
    const piecesWithStamina = units.map((p) => ({
      ...p,
      currentLife: p.unit.lp,
      currentShield: p.unit.shield,
      currentStamina: p.unit.stamina,
    }));
    set({ units: piecesWithStamina });
  },
  setAttackMode: (attackMode) => set({ attackMode }),

  selectPiece: (pieceID) => {
    set({ pieceID });
  },

  startTurn: () => {
    const { currentPlayer, units } = get();
    const updated = units.map((p) =>
      p.player === currentPlayer
        ? { ...p, remainingStamina: p.unit.stamina }
        : p
    );
    set({ units: updated });
  },

  movePiece: (pieceId, targetX, targetY) => {
    const { board, units, currentPlayer } = get();
    const unit = units.find((p) => p.unit.id === pieceId);
    if (!unit) return;
    if (unit.player !== currentPlayer) return;

    const path = findPathToTarget(board, unit, targetX, targetY);
    if (!path) return;

    const totalCost = path.reduce((sum, point) => {
      const cell = board.find((c) => c.x === point.x && c.y === point.y);
      return sum + (cell ? 1 + cell.elevation : 0);
    }, 0);

    if (totalCost > unit.currentStamina) return;

    const updatedPieces = units.map((p) =>
      p.unit.id === pieceId
        ? {
            ...p,
            x: targetX,
            y: targetY,
            currentStamina: p.currentStamina - totalCost,
          }
        : p
    );

    set({ units: updatedPieces });
  },

  endTurn: () => {
    set((state) => ({
      currentPlayer: state.currentPlayer === 1 ? 2 : 1,
    }));
    get().startTurn();
  },

  attackPiece: (targetX: number, targetY: number) => {
    console.log("attackPiece => ", targetX, targetY);
    const { pieceID, units, currentPlayer, attackMode } = get();
    if (!attackMode) return;

    const attacker = units.find((p) => p.unit.id === pieceID);
    if (!attacker) return;
    if (attacker.player !== currentPlayer) return;

    const target = units.find((p) => p.x === targetX && p.y === targetY);
    if (!target) return;
    console.log("target => ", target);
    const weapon = attacker.unit.weapons[0];
    if (!weapon) return;

    const dx = Math.abs(attacker.x - target.x);
    const dy = Math.abs(attacker.y - target.y);
    const distance = Math.max(dx, dy);

    if (distance > weapon.range) return;

    let damage = weapon.damage;
    let newShield = target.currentShield;
    let newLP = target.currentLife;

    if (newShield > 0) {
      const shieldDamage = Math.min(damage, newShield);
      newShield -= shieldDamage;
      damage -= shieldDamage;
    }

    if (damage > 0) {
      newLP -= damage;
    }

    const updatedPieces = units.map((p) => {
      if (p.unit.id === target.unit.id) {
        return {
          ...p,
          currentLife: Math.max(0, newLP),
          currentShield: Math.max(0, newShield),
          unit: { ...p.unit },
        };
      }
      return p;
    });
    console.log("updatedPieces => ", updatedPieces);

    set({ units: updatedPieces, attackMode: false });
  },
}));

export const setInitialState = (initial: {
  board: Cell[];
  units: UnitInstance[];
  pieceID: string | null;
  currentPlayer: 1 | 2;
}) => {
  store.setState({ ...initial });
};

export const useGameStore = () => useStore(store, (s) => s);
