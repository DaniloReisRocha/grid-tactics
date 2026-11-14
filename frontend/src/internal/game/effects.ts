import type { Weapon } from "@/api/models";

export const EffectID = {
  knockback: "knockback",
  shieldBreak: "shieldBreak",
  noAdjacent: "noAdjacentTargets",
  sameLevel: "sameLevelOnly",
} as const;

export function weaponEffectProblem(weapon: Weapon, distance: number, attackerHeight?: number, targetHeight?: number) {
  const has = (id: string) => (weapon.effect ?? []).some((effect) => effect.id === id);
  if (has(EffectID.noAdjacent) && distance <= 1) return "Esta arma não atinge unidades adjacentes.";
  if (has(EffectID.sameLevel)) {
    if (attackerHeight === undefined || targetHeight === undefined) return "Terreno do atacante ou alvo não encontrado.";
    if (attackerHeight !== targetHeight) return "Esta arma exige que atacante e alvo estejam no mesmo nível.";
  }
  return null;
}
