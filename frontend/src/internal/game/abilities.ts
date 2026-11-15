import type { Ability, Cell, Position } from "@/api/models";
import type { SkillToken } from "@/api/skillsLab";
import { isWalkable } from "@/components/ProceduralMap/characterMovement";
import { EffectID } from "./effects";

export function labAbilityProblem(actor: SkillToken | undefined, ability: Ability | undefined, board: readonly Cell[], tokens: readonly SkillToken[], matchUses: Record<string, number>, target?: Position) {
  if (!actor || actor.life <= 0) return "Selecione uma unidade viva.";
  if (!ability || !actor.abilityIDs?.includes(ability.id)) return "Equipe a habilidade na unidade.";
  if (ability.uses === "single" && (actor.abilityUses?.[ability.id] ?? 0) > 0) return "Esta unidade já utilizou a habilidade.";
  const matchCount = Math.max(matchUses[ability.id] ?? 0, tokens.reduce((sum,token) => sum+(token.abilityUses?.[ability.id] ?? 0),0));
  if (ability.maxUsesPerMatch > 0 && matchCount >= ability.maxUsesPerMatch) return "Limite por partida atingido.";
  const has = (id: string) => ability.effect.some((effect) => effect.id === id);
  if (has(EffectID.grantShield) && actor.shield >= ability.shield) return "A unidade já possui o escudo máximo.";
  if (has(EffectID.createBarricade) && ability.maxActivePerMatch > 0 && board.filter((tile) => tile.object === "barricade").length >= ability.maxActivePerMatch) return "Limite de três barricadas ativas atingido.";
  if (!target) return null;
  const from = board.find((tile) => tile.x === actor.x && tile.y === actor.y);
  const to = board.find((tile) => tile.x === target.x && tile.y === target.y);
  if (!from || !to) return "Terreno fora do mapa.";
  const dx = target.x-actor.x, dy = target.y-actor.y;
  const distance = Math.abs(dx)+Math.abs(dy);
  if (dx !== 0 && dy !== 0) return "Escolha um alvo na mesma linha ou coluna.";
  if (distance > ability.range) return "Alvo fora do alcance da habilidade.";
  if (has(EffectID.sameLevel) && from.height !== to.height) return "Escolha um terreno do mesmo nível.";
  const occupied = tokens.some((token) => token.life > 0 && token.x === target.x && token.y === target.y);
  if (has(EffectID.grantShield) && distance !== 0) return "O escudo só pode ser aplicado à própria unidade.";
  if (has(EffectID.createBarricade) && (distance !== 1 || occupied || to.terrain === "water" || to.object)) return "Escolha um terreno adjacente livre e seco.";
  if (has(EffectID.leap)) {
    const move = ability.movement;
    if (!move || occupied || !isWalkable(to)) return "O salto exige um destino acessível e livre.";
    const flat = distance === move.forwardDistance && from.height === to.height;
    const up = distance === move.climbDistance && to.height-from.height === move.climbLevels;
    if (!flat && !up) return "Salte dois terrenos no mesmo nível ou suba um terreno adjacente de um nível acima.";
    if (flat && !board.some((tile) => tile.x === actor.x+dx/2 && tile.y === actor.y+dy/2)) return "O salto não pode atravessar uma lacuna fora do mapa.";
  }
  return null;
}

export function abilityArea(ability: Ability | undefined, center: Position | null, board: readonly Cell[]) {
  if (!ability || !center || !ability.effect.some((effect) => effect.id === EffectID.areaDamageOrShieldBreak)) return [];
  const radius = Math.floor(ability.areaSize/2);
  return board.filter((tile) => Math.abs(tile.x-center.x) <= radius && Math.abs(tile.y-center.y) <= radius);
}
