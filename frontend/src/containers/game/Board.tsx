import { useMemo, useState } from "react";
import ProceduralMap from "@/components/ProceduralMap";
import type { MapTile } from "@/components/ProceduralMap/generator";
import type { MapActor } from "@/components/ProceduralMap/types";
import { useGameSocket } from "@/store/GameSocket";
import { movementOptions, positionKey } from "@/internal/game/pathfinding";
import { playerAppearance } from "@/internal/game/appearance";

export const Board = () => {
  const { id, board, players, turnOrder, unitID, setUnit, movePiece, attackPiece, attackMode,
    accountID, currentTurn, pending, status, connection } = useGameSocket();
  const [hovered, setHovered] = useState<MapTile | null>(null);
  const units = useMemo(() => Object.values(players).flatMap((player) => player.units), [players]);
  const lookup = useMemo(() => new Map(board.map((tile) => [positionKey(tile), tile])), [board]);
  const selected = units.find((unit) => unit.unit.id === unitID);
  const canAct = Boolean(selected && selected.ownerID === accountID && selected.currentLife > 0
    && currentTurn === accountID && status === "playing" && connection === "connected" && !pending);
  const options = useMemo(() => selected && canAct ? movementOptions(board, selected, units) : null, [board, selected, units, canAct]);
  const living = useMemo(() => new Map(units.filter((unit) => unit.currentLife > 0).map((unit) => [positionKey(unit), unit])), [units]);
  const weapon = selected?.unit.weapons[0];
  const attackTargets = new Set<string>();
  if (canAct && attackMode && selected && weapon && selected.currentStamina >= Math.max(1, weapon.staminaPenalty)) {
    for (const unit of living.values()) {
      if (unit.ownerID !== accountID && Math.max(Math.abs(unit.x - selected.x), Math.abs(unit.y - selected.y)) <= weapon.range) {
        attackTargets.add(positionKey(unit));
      }
    }
  }
  const selectable = new Set<string>(living.keys());
  if (canAct && !attackMode) for (const key of options?.paths.keys() ?? []) selectable.add(key);
  for (const key of attackTargets) selectable.add(key);
  const preview = hovered && !attackMode ? options?.paths.get(positionKey(hovered)) : undefined;
  const previewPath = preview && selected ? [{ x: selected.x, y: selected.y }, ...preview] : [];
  const actors = useMemo(() => units.filter((unit) => unit.currentLife > 0).flatMap((unit): MapActor[] => {
    const tile = lookup.get(positionKey(unit));
    if (!tile) return [];
    const playerIndex = Math.max(0, turnOrder.indexOf(unit.ownerID));
    const appearance = playerAppearance(id, playerIndex);
    const path = (unit.path ?? []).map((point) => lookup.get(positionKey(point))).filter((point): point is MapTile => Boolean(point));
    return [{ id: unit.unit.id, tile, path, ...appearance, selected: unit.unit.id === unitID }];
  }), [units, lookup, turnOrder, id, unitID]);
  const width = Math.max(1, ...board.map((cell) => cell.x + 1));
  const height = Math.max(1, ...board.map((cell) => cell.y + 1));

  return (
    <ProceduralMap
      key={id}
      width={width} height={height}
      tiles={board} actors={actors}
      style={{ flex: "1 1 650px", minWidth: 0, height: "max(540px, calc(100vh - 140px))", borderRadius: 12 }}
      interaction={{
        path: previewPath, selectable, attackTargets,
        onHover: setHovered,
        onSelect: (tile) => {
          const key = positionKey(tile);
          const unit = living.get(key);
          if (attackMode && attackTargets.has(key) && selected) {
            attackPiece(selected.unit.id, tile.x, tile.y);
          } else if (unit) {
            setUnit(unit.unit.id);
          } else if (canAct && !attackMode && options?.paths.has(key) && selected) {
            movePiece(selected.unit.id, tile.x, tile.y);
          }
        },
      }}
    />
  );
};
