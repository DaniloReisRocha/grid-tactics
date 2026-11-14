import type { Cell, UnitInstance } from "@/api/models";
import { weaponEffectProblem } from "./effects";

/** Attack allowance is independent of movement stamina; staminaPenalty is ignored. */
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
  const from = board.find((tile) => tile.x === attacker.x && tile.y === attacker.y);
  const to = board.find((tile) => tile.x === target.x && tile.y === target.y);
  const allowed = aligned && distance <= attacker.unit.weapons[0].range
    && !weaponEffectProblem(attacker.unit.weapons[0], distance, from?.height, to?.height);
  if ((target?.x ?? 0) === 5 && (target?.y ?? 0) === 6) {
    console.log("distance => ", distance);
    console.log("range => ", attacker.unit.weapons[0].range);
    console.log("IF => ", allowed);
  }

  return allowed;
}
