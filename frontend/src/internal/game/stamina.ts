import type { Weapon } from "@/api/models";

export function effectiveStamina(base: number, weapons: readonly Pick<Weapon, "staminaPenalty">[]) {
  return Math.max(0, base - weapons.reduce((sum, weapon) => sum + Math.max(0, weapon.staminaPenalty ?? 0), 0));
}
