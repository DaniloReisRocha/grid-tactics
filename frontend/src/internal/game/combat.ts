import type { Cell, UnitInstance } from "@/api/models";
import { directAttackProblem } from "./attackLine";
import { weaponReloading } from "./effects";

/** Attack allowance is independent of stamina; weapon penalties reduce movement capacity only. */
export function canUnitAttack(unit: UnitInstance) {
  return (
    unit.currentLife > 0 &&
    !unit.hasAttacked &&
    unit.unit.weapons.length > 0 &&
    !weaponReloading(unit.unit.weapons[0], unit.turnCount)
  );
}

export function isAttackTarget(
  attacker: UnitInstance,
  target: UnitInstance | undefined,
  board: readonly Cell[] = [],
) {
  if (
    !target ||
    !canUnitAttack(attacker) ||
    target.currentLife <= 0 ||
    target.ownerID === attacker.ownerID
  )
    return false;
  const allowed = !directAttackProblem(
    attacker,
    target,
    attacker.unit.weapons[0],
    board,
  );

  return allowed;
}

export function isBarricadeTarget(
  attacker: UnitInstance,
  tile: Cell,
  board: readonly Cell[],
) {
  return (
    tile.object === "barricade" &&
    canUnitAttack(attacker) &&
    !directAttackProblem(attacker, tile, attacker.unit.weapons[0], board)
  );
}
