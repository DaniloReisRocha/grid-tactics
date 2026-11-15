import type { Cell, UnitInstance } from "@/api/models";
import { directAttackProblem } from "./attackLine";

/** Attack allowance is independent of stamina; weapon penalties reduce movement capacity only. */
export function canUnitAttack(unit: UnitInstance) {
  return (
    unit.currentLife > 0 && !unit.hasAttacked && unit.unit.weapons.length > 0
  );
}

export function isAttackTarget(
  attacker: UnitInstance,
  target: UnitInstance | undefined,
  board: readonly Cell[] = [],
) {
  if ((target?.x ?? 0) === 5 && (target?.y ?? 0) === 6) {
    console.log("------------------------------------------");
    console.log("targuet => ", target);
    console.log("No targuet => ", !target);
    console.log("No canUnitAttack => ", !canUnitAttack(attacker));
    console.log("No currentLife => ", (target?.currentLife ?? 0) <= 0);
    console.log("friend fire => ", target?.ownerID === attacker.ownerID);
  }

  if (
    !target ||
    !canUnitAttack(attacker) ||
    target.currentLife <= 0 ||
    target.ownerID === attacker.ownerID
  )
    return false;
  const dx = Math.abs(target.x - attacker.x);
  const dy = Math.abs(target.y - attacker.y);
  const aligned = dx === 0 || dy === 0;
  const distance = dx + dy;
  const allowed = aligned && distance <= attacker.unit.weapons[0].range
    && !directAttackProblem(attacker,target,attacker.unit.weapons[0],board);
  if ((target?.x ?? 0) === 5 && (target?.y ?? 0) === 6) {
    console.log("distance => ", distance);
    console.log("range => ", attacker.unit.weapons[0].range);
    console.log("IF => ", allowed);
  }

  return allowed;
}

export function isBarricadeTarget(attacker: UnitInstance,tile: Cell,board: readonly Cell[]) {
  return tile.object === "barricade" && canUnitAttack(attacker)
    && !directAttackProblem(attacker,tile,attacker.unit.weapons[0],board);
}
