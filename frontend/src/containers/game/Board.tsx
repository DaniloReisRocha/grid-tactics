import { GameStatus, ConnectionStatus } from "@/api/status";
import { useMemo, useState } from "react";
import ProceduralMap from "@/components/ProceduralMap";
import type { MapTile } from "@/components/ProceduralMap/generator";
import type { MapActor } from "@/components/ProceduralMap/types";
import { useGameSocket } from "@/store/GameSocket";
import { movementOptions, positionKey } from "@/internal/game/pathfinding";
import { playerAppearance } from "@/internal/game/appearance";
import { canUnitAttack, isAttackTarget, isBarricadeTarget } from "@/internal/game/combat";
import { labAbilityProblem,combatSkillToken,abilityArea } from "@/internal/game/abilities";
import { useGameOrbs } from "./useGameOrbs";
import { EffectID } from "@/internal/game/effects";
import TurnTimer from "./TurnTimer";
import { attackableArea } from "@/internal/game/attackLine";

export const Board = () => {
  const { id, board, players, orbs, grenadeExplosions, finishGrenadeExplosion, turnOrder, unitID, setUnit, movePiece, attackPiece, attackMode,
    accountID, currentTurn, currentUnitID, pending, status, connection,skillMode,activateAbility,matchAbilityUses } = useGameSocket();
  const [hovered, setHovered] = useState<MapTile | null>(null);
  const units = useMemo(() => Object.values(players).flatMap((player) => player.units), [players]);
  const orbView = useGameOrbs(orbs, units);
  const lookup = useMemo(() => new Map(board.map((tile) => [positionKey(tile), tile])), [board]);
  const selected = units.find((unit) => unit.unit.id === unitID);
  const skillTokens=useMemo(()=>units.map(combatSkillToken),[units]);
  const selectedToken=skillTokens.find((unit)=>unit.id===unitID);
  const activeSkill=selected?.unit.abilities?.find((item)=>item.id===skillMode);
  const canAct = Boolean(selected && selected.ownerID === accountID && selected.currentLife > 0
    && selected.unit.id === currentUnitID
    && currentTurn === accountID && status === GameStatus.playing && connection === ConnectionStatus.connected && !pending);
  const options = useMemo(() => selected && canAct && !skillMode ? movementOptions(board, selected, units) : null, [board, selected, units, canAct,skillMode]);
  const living = useMemo(() => new Map(units.filter((unit) => unit.currentLife > 0).map((unit) => [positionKey(unit), unit])), [units]);
  const attackTargets = new Set<string>();
  let attackArea = new Set<string>();
  const barricadeTargets = new Set<string>();
  if (canAct && attackMode && selected && canUnitAttack(selected)) {
    attackArea = attackableArea(selected, selected.unit.weapons[0], board);
    for (const unit of living.values()) {
      if (isAttackTarget(selected, unit, board)) {
        attackTargets.add(positionKey(unit));
      }
    }
    for (const tile of board) if (isBarricadeTarget(selected,tile,board)) barricadeTargets.add(positionKey(tile));
  }
  const selectable = new Set<string>(attackMode || skillMode ? [] : living.keys());
  if (canAct&&skillMode) for(const tile of board) {
    const orbBlocksStructure = activeSkill?.effect.some((effect) => effect.id === EffectID.createBarricade)
      && orbs.some((orb) => !orb.triggered && orb.x === tile.x && orb.y === tile.y);
    if(!orbBlocksStructure&&!labAbilityProblem(selectedToken,activeSkill,board,skillTokens,matchAbilityUses??{},tile))selectable.add(positionKey(tile));
  }
  if (canAct && !attackMode) for (const key of options?.paths.keys() ?? []) selectable.add(key);
  for (const key of attackTargets) selectable.add(key);
  for (const key of barricadeTargets) selectable.add(key);
  const preview = hovered && !attackMode && !skillMode ? options?.paths.get(positionKey(hovered)) : undefined;
  const previewPath = preview && selected ? [{ x: selected.x, y: selected.y }, ...preview] : [];
  const actors = useMemo(() => units.flatMap((unit): MapActor[] => {
    const tile = lookup.get(positionKey(unit));
    if (!tile) return [];
    const playerIndex = Math.max(0, turnOrder.indexOf(unit.ownerID));
    const appearance = playerAppearance(id, playerIndex);
    const path = (unit.path ?? []).map((point) => lookup.get(positionKey(point))).filter((point): point is MapTile => Boolean(point));
    const defeated = unit.currentLife <= 0;
    return [{
      id: unit.unit.id, tile, path,
      life: unit.currentLife, shield: unit.currentShield,
      movementMode:unit.movementMode,
      ...appearance, defeated, selected: !defeated && unit.unit.id === unitID,
      activeTurn: !defeated && unit.unit.id === currentUnitID,
    }];
  }), [units, lookup, turnOrder, id, unitID, currentUnitID]);
  const width = Math.max(1, ...board.map((cell) => cell.x + 1));
  const height = Math.max(1, ...board.map((cell) => cell.y + 1));

  return (
    <ProceduralMap
      key={id}
      width={width} height={height}
      tiles={board} actors={actors}
      orbs={orbView.orbs} onOrbComplete={orbView.onOrbComplete}
      explosions={grenadeExplosions} onExplosionComplete={finishGrenadeExplosion}
      topLeftOverlay={<TurnTimer />}
      style={{ flex: "1 1 650px", minWidth: 0, height: "max(540px, calc(100vh - 140px))", borderRadius: 12 }}
      interaction={{
        path: previewPath, selectable, attackTargets, attackArea: attackMode ? attackArea : undefined,
        area:skillMode&&hovered&&selectable.has(positionKey(hovered))?abilityArea(activeSkill,hovered,board):[],
        onHover: setHovered,
        onActorTileEnter: orbView.onActorTileEnter,
        onSelect: (tile) => {
          const key = positionKey(tile);
          const unit = living.get(key);
          if(skillMode){if(selected&&canAct&&selectable.has(key))activateAbility(selected.unit.id,skillMode,tile.x,tile.y);return;}
          if (attackMode) {
            if ((attackTargets.has(key) || barricadeTargets.has(key)) && selected) {
              attackPiece(selected.unit.id, tile.x, tile.y);
            }
            return;
          }
          if (unit) {
            setUnit(unit.unit.id);
          } else if (canAct && !attackMode && options?.paths.has(key) && selected) {
            movePiece(selected.unit.id, tile.x, tile.y);
          }
        },
      }}
    />
  );
};
