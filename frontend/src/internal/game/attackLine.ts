import type { Cell, Position, Weapon } from "@/api/models";
import { weaponEffectProblem } from "./effects";

export function blocksDirectAttack(tile: Cell) {
  return tile.object === "house" || tile.object === "ruin" || tile.object === "barricade";
}

export function directAttackProblem(from: Position, target: Position, weapon: Weapon, board: readonly Cell[]) {
  const dx=target.x-from.x,dy=target.y-from.y;
  if ((dx===0&&dy===0)||(dx!==0&&dy!==0)) return "O alvo precisa estar na mesma linha ou coluna do atacante.";
  const distance=Math.abs(dx)+Math.abs(dy);
  if (distance>weapon.range) return "Alvo fora do alcance da arma.";
  const cells=new Map(board.map((tile)=>[`${tile.x},${tile.y}`,tile]));
  const source=cells.get(`${from.x},${from.y}`),destination=cells.get(`${target.x},${target.y}`);
  const effectProblem=weaponEffectProblem(weapon,distance,source?.height,destination?.height);
  if (effectProblem) return effectProblem;
  // Retain geometry-only checks for callers without a map; game actions supply it.
  if (!board.length) return null;
  const sx=Math.sign(dx),sy=Math.sign(dy);
  for (let x=from.x+sx,y=from.y+sy;;x+=sx,y+=sy) {
    const tile=cells.get(`${x},${y}`);
    if (!tile) return "Terreno fora do mapa.";
    if (x===target.x&&y===target.y) return !blocksDirectAttack(tile)||tile.object==="barricade" ? null : "Uma estrutura bloqueia a linha de ataque.";
    if (blocksDirectAttack(tile)) return "Uma estrutura bloqueia a linha de ataque.";
  }
}
