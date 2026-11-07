import { useEffect, useMemo } from "react";
import { setInitialState } from "@/store/game";
import type { Cell, PiecePoint, Piece, Weapon } from "@/api/models";

function seededRandom(seed: string): () => number {
  let h = 2166136261 >>> 0;
  for (let i = 0; i < seed.length; i++) {
    h ^= seed.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }

  return () => {
    h += h << 13;
    h ^= h >>> 7;
    h += h << 3;
    h ^= h >>> 17;
    h += h << 5;
    return (h >>> 0) / 4294967296;
  };
}

const generateBoard = (sid: string): Cell[] => {
  const rand = seededRandom(sid);
  const cells: Cell[] = [];

  for (let y = 0; y < 10; y++) {
    for (let x = 0; x < 10; x++) {
      cells.push({ x, y, elevation: 0 });
    }
  }

  const elevationCount = Math.floor(rand() * 11) + 15;
  const used = new Set<string>();
  while (used.size < elevationCount) {
    const x = Math.floor(rand() * 10);
    const y = Math.floor(rand() * 10);
    const key = `${x}-${y}`;
    if (used.has(key)) continue;
    used.add(key);
    const i = y * 10 + x;
    const elevation = rand() < 0.5 ? 1 : 2;
    cells[i].elevation = elevation;
  }

  return cells;
};

const generateInitialPieces = (sid: string, parts: Piece[][]): PiecePoint[] => {
  const rand = seededRandom(sid);
  const used = new Set<string>();
  const pieces: PiecePoint[] = [];

  parts.forEach((player, index) => {
    player.forEach((piece) => {
      let x = 0;
      let y = 0;
      let attempts = 0;
      let positionFound = false;

      do {
        x = Math.floor(rand() * 10);
        y = Math.floor(rand() * 10);
        attempts++;

        const tooClose = Array.from(used).some((pos) => {
          const [ux, uy] = pos.split("-").map(Number);
          const dx = Math.abs(ux - x);
          const dy = Math.abs(uy - y);
          return dx <= 1 && dy <= 1;
        });

        if (!used.has(`${x}-${y}`) && !tooClose) {
          positionFound = true;
          break;
        }
      } while (attempts < 1000);

      if (!positionFound) {
        outer: for (let yy = 0; yy < 10; yy++) {
          for (let xx = 0; xx < 10; xx++) {
            if (!used.has(`${xx}-${yy}`)) {
              x = xx;
              y = yy;
              break outer;
            }
          }
        }
      }

      used.add(`${x}-${y}`);
      pieces.push({
        x,
        y,
        piece,
        player: index + 1,
        currentLife: piece.lp,
        currentShield: piece.shield,
        currentStamina: piece.stamina,
      });
    });
  });

  return pieces;
};

export const PISTOL: Weapon = {
  name: "Pistola",
  icon: "🔫",
  damage: 4,
  range: 2,
  staminaPenalty: 0,
  toolPenalty: 0,
  description: "Item base do jogo",
};

const GamePlayers = [
  [
    {
      icon: "🦸‍♂️",
      id: "11",
      lp: 3,
      shield: 3,
      stamina: 100,
      weapons: [PISTOL],
      tools: [],
    },
    {
      icon: "🦸‍♀️",
      id: "12",
      lp: 3,
      shield: 3,
      stamina: 4,
      weapons: [PISTOL],
      tools: [],
    },
    {
      icon: "🧙‍♂️",
      id: "13",
      lp: 3,
      shield: 3,
      stamina: 4,
      weapons: [PISTOL],
      tools: [],
    },
  ],
  [
    {
      icon: "🦹‍♀️",
      id: "21",
      lp: 3,
      shield: 3,
      stamina: 4,
      weapons: [PISTOL],
      tools: [],
    },
    {
      icon: "🦹‍♂️",
      id: "22",
      lp: 3,
      shield: 3,
      stamina: 4,
      weapons: [PISTOL],
      tools: [],
    },
    {
      icon: "🧛",
      id: "23",
      lp: 3,
      shield: 3,
      stamina: 4,
      weapons: [PISTOL],
      tools: [],
    },
  ],
];

export const useStartGame = () => {
  const game = useMemo(() => {
    const sid = `${Math.floor(Math.random() * 10000)}`;
    const board = generateBoard(sid);
    const pieces = generateInitialPieces(sid, GamePlayers);
    return { board, pieces };
  }, []);

  useEffect(() => {
    const { board, pieces } = game;
    setInitialState({
      board,
      pieces,
      pieceID: null,
      currentPlayer: 1,
    });
  }, [JSON.stringify(game)]);
};
