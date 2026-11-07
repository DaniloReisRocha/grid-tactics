import { type FC, useCallback, useMemo, useState } from "react";
import { useGameSocket } from "@/store/GameSocket";
import type { Cell, UnitInstance } from "@/api/models";
import "@/styles/game/Board.css";

export function getReachableCells(
  board: Cell[],
  unit: UnitInstance
): { x: number; y: number }[] {
  const reachable: { x: number; y: number }[] = [];
  const queue: { x: number; y: number; cost: number }[] = [
    { x: unit.x, y: unit.y, cost: 0 },
  ];
  const visited = new Set<string>();

  while (queue.length > 0) {
    const { x, y, cost } = queue.shift()!;
    const key = `${x}-${y}`;

    if (visited.has(key)) continue;
    visited.add(key);

    const cell = board.find((c) => c.x === x && c.y === y);
    if (!cell) continue;

    if (cost <= unit.currentStamina) {
      reachable.push({ x, y });

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
        const nx = x + dx;
        const ny = y + dy;
        const nextCell = board.find((c) => c.x === nx && c.y === ny);
        if (!nextCell) continue;

        const moveCost = 1 + nextCell.elevation;
        queue.push({ x: nx, y: ny, cost: cost + moveCost });
      }
    }
  }

  return reachable.filter((pos) => !(pos.x === unit.x && pos.y === unit.y));
}

function hasTargetInRange(
  x: number,
  y: number,
  player: number,
  selectedPiece?: UnitInstance
): boolean {
  if (!selectedPiece || selectedPiece?.player === player || player === 0) {
    return false;
  }
  const weapon = selectedPiece?.unit?.weapons?.[0];
  const dx = Math.abs(x - selectedPiece.x);
  const dy = Math.abs(y - selectedPiece.y);
  return Math.max(dx, dy) <= weapon.range;
}

export function getLinePath(
  startX: number,
  startY: number,
  endX: number,
  endY: number
): { x: number; y: number }[] {
  const path: { x: number; y: number }[] = [];
  const dx = endX - startX;
  const dy = endY - startY;
  const steps = Math.max(Math.abs(dx), Math.abs(dy));

  for (let i = 1; i <= steps; i++) {
    const x = Math.round(startX + (dx * i) / steps);
    const y = Math.round(startY + (dy * i) / steps);
    path.push({ x, y });
  }

  return path;
}

export const Board: FC = () => {
  const { board } = useGameSocket();
  const [hoverPath, setHoverPath] = useState<{ x: number; y: number }[]>([]);

  return (
    <div className="board">
      {board.map((cell) => (
        <RenderPieceAt
          key={`${cell.x}-${cell.y}`}
          x={cell.x}
          y={cell.y}
          elevation={cell.elevation}
          hoverPath={hoverPath}
          setHoverPath={setHoverPath}
        />
      ))}
    </div>
  );
};

type RenderPieceAtProps = {
  x: number;
  y: number;
  elevation: number;
  hoverPath: { x: number; y: number }[];
  setHoverPath: (path: { x: number; y: number }[]) => void;
};

const RenderPieceAt = (props: RenderPieceAtProps) => {
  const { elevation, x, y, hoverPath, setHoverPath } = props;
  const {
    board,
    // units,
    // pieceID,
    unitID,
    setUnit,
    players,
    movePiece,
    // attackMode,
    // attackPiece,
  } = useGameSocket();
  const attackMode = false;
  const units = Object.values(players).reduce((p1, p2) => {
    return p1.concat(p2?.units ?? []);
  }, [] as UnitInstance[]);
  console.log("units => ", units);
  console.log("players => ", players);

  const attackPiece = (x: number, y: number) => {};

  const point = units.find((p) => p.x === x && p.y === y);
  const selectedPiece = units.find((p) => p.unit.id === unitID);
  const isSelected = selectedPiece?.unit?.id === point?.unit?.id;

  const reachableCells = useMemo(() => {
    if (!selectedPiece) return [];
    return getReachableCells(board, selectedPiece);
  }, [selectedPiece, board]);

  const isReachable = useCallback(() => {
    return reachableCells.some((cell) => cell.x === x && cell.y === y);
  }, [x, y, reachableCells]);

  const isHoverPath = useCallback(() => {
    return hoverPath.some((p) => p.x === x && p.y === y);
  }, [x, y, hoverPath]);

  const handleCellMouseLeave = useCallback(() => {
    setHoverPath([]);
  }, [x, y]);

  const handleCellMouseEnter = useCallback(() => {
    if (!selectedPiece || !isReachable()) {
      setHoverPath([]);
      return;
    }
    const path = getLinePath(selectedPiece.x, selectedPiece.y, x, y);
    setHoverPath(path);
  }, [x, y, isReachable, setHoverPath]);

  return (
    <div
      className={`tile elevation-${elevation}
      ${isReachable() ? "reachable" : ""}
      ${isHoverPath() ? "hover-line" : ""}
      ${
        attackMode && hasTargetInRange(x, y, point?.player ?? 0, selectedPiece)
          ? "attack-target"
          : ""
      }`}
      onClick={(e) => {
        e.preventDefault();
        e.stopPropagation();
        const pieceId = selectedPiece?.unit?.id;
        if (point && attackMode) {
          attackPiece(x, y);
          return;
        }
        if (point && !attackMode) {
          setUnit(point.unit.id);
          return;
        }
        if (pieceId && !attackMode) {
          console.log(pieceId, x, y);
          movePiece(pieceId, x, y);
        }
      }}
      onMouseEnter={handleCellMouseEnter}
      onMouseLeave={handleCellMouseLeave}
    >
      <span className="elevation-label">{elevation}</span>
      {point ? (
        <div
          className={`unit ${`player${point.player}`} ${
            isSelected ? "selected" : ""
          }`}
        >
          {point.currentLife === 0 ? "☠️" : point.unit.icon}
        </div>
      ) : null}
    </div>
  );
};
