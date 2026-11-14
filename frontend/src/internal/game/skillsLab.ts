import type { Cell, UnitInstance, Weapon } from "@/api/models";
import type { SkillToken } from "@/api/skillsLab";
import { isWalkable } from "@/components/ProceduralMap/characterMovement";

export function labAttackProblem(attacker: SkillToken, target: SkillToken | undefined, board: Cell[], weapons: Weapon[]) {
  if (attacker.life <= 0) return "Escolha um atacante vivo.";
  if (!target || target.life <= 0) return "Clique em um token vivo da outra equipe.";
  if (attacker.team === target.team) return "Escolha um alvo da outra equipe.";
  const weapon = weapons.find((item) => item.id === attacker.weaponID);
  if (!weapon) return "Equipe uma arma no atacante.";
  const dx = Math.abs(target.x - attacker.x);
  const dy = Math.abs(target.y - attacker.y);
  if (dx !== 0 && dy !== 0) return "O alvo precisa estar na mesma linha ou coluna do atacante.";
  const distance = dx + dy;
  if (distance > weapon.range) return "Alvo fora do alcance da arma.";
  if (weapon.id === "sniper" && distance <= 1) return "O rifle sniper não atinge unidades adjacentes.";
  const from = board.find((tile) => tile.x === attacker.x && tile.y === attacker.y);
  const to = board.find((tile) => tile.x === target.x && tile.y === target.y);
  if (!from || !to) return "Token fora do mapa.";
  if ((weapon.id === "katana" || weapon.id === "warHammer") && from.height !== to.height) {
    return "Esta arma exige que atacante e alvo estejam no mesmo nível.";
  }
  return null;
}

// Adapt creative tokens to the same movement/path preview used in a real game.
export function labMovementUnit(token: SkillToken, weapons: Weapon[]): UnitInstance {
  return {
    x: token.x, y: token.y, player: token.team, ownerID: String(token.team),
    currentLife: token.life, currentShield: token.shield, currentStamina: 999,
    hasAttacked: false, path: token.path,
    unit: { id: token.id, lp: token.maxLife, shield: token.maxShield, stamina: 999,
      icon: "", weapons: weapons.filter((weapon) => weapon.id === token.weaponID), tools: [] },
  };
}

export function starterTokens(board: Cell[], weapons: Weapon[]): SkillToken[] {
  const candidates = board.filter(isWalkable).sort((a, b) =>
    Math.abs(a.x - 7) + Math.abs(a.y - 7) - Math.abs(b.x - 7) - Math.abs(b.y - 7) || a.y - b.y || a.x - b.x);
  const anchor = candidates[0];
  const targets = anchor ? candidates.filter((tile) => tile !== anchor
    && tile.height === anchor.height
    && (tile.x === anchor.x || tile.y === anchor.y)
    && Math.abs(tile.x - anchor.x) + Math.abs(tile.y - anchor.y) <= 2) : [];
  const chosen = anchor ? [anchor, ...targets.slice(0, 2)] : [];
  return chosen.map((tile, index) => ({
    id: `starter-${index + 1}`, x: tile.x, y: tile.y, team: index === 0 ? 0 : 1,
    model: index as 0 | 1 | 2, life: 6, shield: 3, maxLife: 6, maxShield: 3,
    weaponID: weapons[0]?.id ?? "pistol", path: [],
  }));
}
