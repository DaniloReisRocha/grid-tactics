import { useMemo, useState } from "react";
import { useGameSocket } from "@/store/GameSocket";
import { movementOptions, positionKey } from "@/internal/game/pathfinding";
import "@/styles/game/Board.css";

export const Board = () => {
  const { board, players, unitID, setUnit, movePiece, attackPiece, attackMode, accountID, currentTurn, pending, status, connection } = useGameSocket();
  const [hovered, setHovered] = useState<string | null>(null);
  const units = useMemo(() => Object.values(players).flatMap((player) => player.units), [players]);
  const selected = units.find((unit) => unit.unit.id === unitID);
  const canAct = Boolean(selected && selected.ownerID === accountID && selected.currentLife > 0
    && currentTurn === accountID && status === "playing" && connection === "connected" && !pending);
  const options = useMemo(() => selected && canAct ? movementOptions(board, selected, units) : null, [board, selected, units, canAct]);
  const path = new Set((hovered && !attackMode ? options?.paths.get(hovered) ?? [] : []).map(positionKey));
  const weapon = selected?.unit.weapons[0];
  const columns = Math.max(1, ...board.map((cell) => cell.x + 1));
  return (
    <div className="board" style={{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }}
      onMouseLeave={() => setHovered(null)} aria-label="Tabuleiro da partida">
      {board.map((cell) => {
        const key = positionKey(cell);
        const living = units.find((unit) => unit.x === cell.x && unit.y === cell.y && unit.currentLife > 0);
        const corpse = units.find((unit) => unit.x === cell.x && unit.y === cell.y && unit.currentLife <= 0);
        const reachable = canAct && !attackMode && !living && options?.paths.has(key);
        const inRange = Boolean(attackMode && canAct && living && living.ownerID !== accountID && weapon
          && selected && Math.max(Math.abs(cell.x - selected.x), Math.abs(cell.y - selected.y)) <= weapon.range
          && selected.currentStamina >= Math.max(1, weapon.staminaPenalty));
        return (
          <button type="button" key={key}
            className={`tile elevation-${cell.elevation} ${reachable ? "reachable" : ""} ${path.has(key) ? "hover-line" : ""} ${inRange ? "attack-target" : ""}`}
            onMouseEnter={() => setHovered(key)} onFocus={() => setHovered(key)}
            title={`(${cell.x}, ${cell.y}) · Elevação ${cell.elevation}${reachable ? ` · Custo ${options?.costs.get(key)}` : ""}`}
            onClick={() => {
              if (attackMode && inRange && selected) { attackPiece(selected.unit.id, cell.x, cell.y); return; }
              if (living) { setUnit(living.unit.id); return; }
              if (reachable && selected) movePiece(selected.unit.id, cell.x, cell.y);
            }}>
            <span className="elevation-label">{cell.elevation}</span>
            {living || corpse ? <span className={`unit player${(living ?? corpse)?.player} ${living?.unit.id === unitID ? "selected" : ""}`}>
              {living ? living.unit.icon : "☠️"}
            </span> : null}
          </button>
        );
      })}
    </div>
  );
};
