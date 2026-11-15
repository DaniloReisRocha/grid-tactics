import type { Ability, Weapon } from "./models";
import { backendURL, backendHeaders } from "./backendURL";

export interface UnitLoadout { weaponID: string; abilityIDs: string[] }
export interface TeamOptions {
  weapons: Weapon[]; abilities: Ability[]; unitCount: number;
  baseLife: number; shieldCapacity: number; baseStamina: number;
  itemSlots: number; abilitySlots: number;
}

export async function loadTeamOptions(signal: AbortSignal): Promise<TeamOptions> {
  let response: Response;
  const url = backendURL("/api/team-options");
  try { response = await fetch(url, { signal, headers: backendHeaders(url) }); }
  catch (error) {
    if (signal.aborted) throw error;
    throw new Error("Não foi possível carregar os equipamentos. Verifique se o backend está atualizado e em execução.");
  }
  if (response.headers.get("Content-Type")?.includes("text/html")) throw new Error("A API devolveu uma página HTML. Confira se o túnel aponta para o backend e se o servidor foi reiniciado.");
  const data = await response.json().catch(() => null);
  if (!response.ok || !data?.weapons || !data?.abilities) throw new Error(data?.error ?? "Não foi possível carregar o catálogo de equipamentos.");
  return data;
}
