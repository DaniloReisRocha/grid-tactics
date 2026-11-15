import type { TeamOptions, UnitLoadout } from "@/api/team";

export const defaultTeam = (): UnitLoadout[] => Array.from({ length: 3 }, () => ({ weaponID: "pistol", abilityIDs: [] }));
type TeamStorage = Pick<Storage, "getItem" | "setItem">;
const matchKey = (session: string, account: string) => `grid-tactics.team:${encodeURIComponent(session)}:${account.toLowerCase()}`;

export function isTeam(value: unknown): value is UnitLoadout[] {
  return Array.isArray(value) && value.length === 3 && value.every((entry: unknown) => {
    if (!entry || typeof entry !== "object") return false;
    const unit = entry as Partial<UnitLoadout>;
    return typeof unit.weaponID === "string" && !!unit.weaponID && Array.isArray(unit.abilityIDs)
      && unit.abilityIDs.length <= 2 && unit.abilityIDs.every((id: unknown) => typeof id === "string");
  });
}

export function loadTeamDraft(storage: TeamStorage = sessionStorage) {
  try { const value: unknown = JSON.parse(storage.getItem("grid-tactics.team.draft") ?? "null"); return isTeam(value) ? value : defaultTeam(); }
  catch { return defaultTeam(); }
}

export function readMatchTeam(session: string, account: string, storage: TeamStorage = sessionStorage) {
  try { const value: unknown = JSON.parse(storage.getItem(matchKey(session, account)) ?? "null"); return isTeam(value) ? value : undefined; }
  catch { return undefined; }
}

export function saveMatchTeam(session: string, account: string, team: UnitLoadout[], storage: TeamStorage = sessionStorage) {
  storage.setItem(matchKey(session, account), JSON.stringify(team));
  storage.setItem("grid-tactics.team.draft", JSON.stringify(team));
}

export function abilitySlotCount(weaponPenalty: number, baseSlots: number) { return Math.max(0, baseSlots - Math.max(0, weaponPenalty)); }

export function validateTeam(team: UnitLoadout[], options: TeamOptions): string | null {
  if (team.length !== options.unitCount) return `Monte uma equipe com ${options.unitCount} unidades.`;
  for (const [index, unit] of team.entries()) {
    const weapon = options.weapons.find((entry) => entry.id === unit.weaponID);
    if (!weapon) return `Unidade ${index + 1}: selecione uma arma válida.`;
    const equipped = unit.abilityIDs.filter(Boolean);
    if (equipped.length > abilitySlotCount(weapon.toolPenalty, options.abilitySlots)) return `Unidade ${index + 1}: a arma reduz os slots disponíveis de habilidades.`;
    if (new Set(equipped).size !== equipped.length || equipped.some((id) => !options.abilities.some((entry) => entry.id === id))) return `Unidade ${index + 1}: escolha habilidades válidas, sem repetir.`;
  }
  return null;
}
