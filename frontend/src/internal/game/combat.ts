import type { UnitInstance } from "@/api/models";

/** Attack allowance is independent of movement stamina; staminaPenalty is ignored. */
export function canUnitAttack(unit: UnitInstance) {
  return (
    unit.currentLife > 0 && !unit.hasAttacked && unit.unit.weapons.length > 0
  );
}

export function isAttackTarget(
  attacker: UnitInstance,
  target: UnitInstance | undefined,
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
  const distance = Math.max(
    Math.abs(target.x - attacker.x),
    Math.abs(target.y - attacker.y),
  );
  if ((target?.x ?? 0) === 5 && (target?.y ?? 0) === 6) {
    console.log("distance => ", distance);
    console.log("range => ", attacker.unit.weapons[0].range);
    console.log("IF => ", distance <= attacker.unit.weapons[0].range);
  }

  return distance <= attacker.unit.weapons[0].range;
}
