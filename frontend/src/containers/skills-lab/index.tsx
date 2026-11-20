import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import ProceduralMap from "@/components/ProceduralMap";
import { isWalkable } from "@/components/ProceduralMap/characterMovement";
import type { MapTile } from "@/components/ProceduralMap/generator";
import type { CharacterModel, MapActor, MapOrb, MapExplosion } from "@/components/ProceduralMap/types";
import { attackInLab, activateAbilityInLab, loadLab, type LabConfig, type LabLayout, type SkillToken } from "@/api/skillsLab";
import { movementOptions, positionKey } from "@/internal/game/pathfinding";
import { labAttackProblem, labMovementUnit, labStamina, starterTokens } from "@/internal/game/skillsLab";
import { effectiveStamina } from "@/internal/game/stamina";
import { labAbilityProblem, abilityArea, applyPassiveAbilities, restoreOrbAbilities } from "@/internal/game/abilities";
import { EffectID, weaponReloading } from "@/internal/game/effects";
import AbilitySelector from "./AbilitySelector";
import "./styles.css";
import { attackableArea } from "@/internal/game/attackLine";

type Tool = "select" | "add" | "attack" | "remove" | "paint" | "skill" | "orb";
const TOOLS: { id: Tool; label: string; hint: string }[] = [
  { id: "select", label: "Selecionar / mover", hint: "Selecione um token. Passe o mouse para ver a rota e clique em um slot para mover." },
  { id: "add", label: "+ Token", hint: "Escolha equipe, personagem e arma. Clique em um slot livre para adicionar." },
  { id: "orb", label: "+ Orb", hint: "Adicione uma orb em um terreno livre. Ao passar por ela, a unidade recupera os usos das habilidades e os efeitos passivos." },
  { id: "attack", label: "Testar ataque", hint: "Clique em um alvo da outra equipe. Os alvos válidos estão marcados em vermelho." },
  { id: "remove", label: "Remover", hint: "Clique em um token, caveira ou orb para remover do cenário." },
  { id: "paint", label: "Editar terreno", hint: "Escolha o terreno e a altura. Clique em um slot sem tokens para pintar." },
];
const TEAMS = ["#68d2ff", "#ffcc66"];
const errorText = (error: unknown) => error instanceof Error ? error.message : "Não foi possível executar a ação.";

export default function SkillsLab() {
  const [config, setConfig] = useState<LabConfig | null>(null);
  const [tokens, setTokens] = useState<SkillToken[]>([]);
  const [orbs, setOrbs] = useState<MapOrb[]>([]);
  const [explosions, setExplosions] = useState<MapExplosion[]>([]);
  const collectedOrbs = useRef(new Set<string>());
  const orbRevision = useRef(0);
  const unitOrbRevision = useRef(new Map<string, number>());
  const [selectedID, setSelectedID] = useState<string | null>(null);
  const [tool, setTool] = useState<Tool>("select");
  const [hovered, setHovered] = useState<MapTile | null>(null);
  const [seed, setSeed] = useState("skills-lab");
  const [layout, setLayout] = useState<LabLayout>("arena");
  const [team, setTeam] = useState(0);
  const [model, setModel] = useState<CharacterModel>(0);
  const [weaponID, setWeaponID] = useState("pistol");
  const [abilityIDs, setAbilityIDs] = useState<string[]>([]);
  const [activeAbilityID, setActiveAbilityID] = useState<string | null>(null);
  const [matchAbilityUses, setMatchAbilityUses] = useState<Record<string,number>>({});
  const [life, setLife] = useState(6);
  const [shield, setShield] = useState(3);
  const [terrain, setTerrain] = useState<"grass" | "dirt" | "water" | "house" | "tree" | "rock">("grass");
  const [height, setHeight] = useState(0);
  const [pending, setPending] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [log, setLog] = useState<string[]>([]);
  const [version, setVersion] = useState(0);
  const activeRequest = useRef<AbortController | null>(null);
  const busy = useRef(false);
  const initial = useRef<{ config: LabConfig; tokens: SkillToken[] } | null>(null);

  const loadScenario = useCallback(async (sceneSeed: string, sceneLayout: LabLayout) => {
    activeRequest.current?.abort();
    const controller = new AbortController();
    activeRequest.current = controller;
    busy.current = true;
    setPending(true); setError(null);
    try {
      const data = await loadLab(sceneSeed, sceneLayout, controller.signal);
      if (controller.signal.aborted) return;
      const starters = starterTokens(data.board, data.weapons);
      initial.current = structuredClone({ config: data, tokens: starters });
      setConfig(data); setTokens(starters); setSelectedID(starters[0]?.id ?? null);
      setOrbs([]);
      setExplosions([]);
      collectedOrbs.current.clear(); unitOrbRevision.current.clear(); orbRevision.current = 0;
      setWeaponID(data.weapons[0]?.id ?? "pistol"); setTool("select");
      setAbilityIDs([]);
      setActiveAbilityID(null); setMatchAbilityUses({});
      setHovered(null); setVersion((value) => value + 1); setLog(["Cenário pronto. Ataques livres; movimento respeita a estamina e as penalidades da arma."]);
    } catch (cause) { if (!controller.signal.aborted) setError(errorText(cause)); }
    finally {
      if (activeRequest.current === controller) { busy.current = false; setPending(false); }
    }
  }, []);

  useEffect(() => {
    void loadScenario("skills-lab", "arena");
    return () => { activeRequest.current?.abort(); };
  }, [loadScenario]);

  const board = useMemo(() => config?.board ?? [], [config]);
  const weapons = useMemo(() => config?.weapons ?? [], [config]);
  const abilities = config?.abilities ?? [];
  const triggerOrb = (tile: MapTile, actorID: string) => {
    const orb = orbs.find((item) => !item.triggered && item.x === tile.x && item.y === tile.y && !collectedOrbs.current.has(item.id));
    if (!orb) return;
    collectedOrbs.current.add(orb.id);
    unitOrbRevision.current.set(actorID, ++orbRevision.current);
    setOrbs((items) => items.map((item) => item.id === orb.id ? {...item, triggered: true} : item));
    setTokens((items) => items.map((token) => token.id === actorID ? restoreOrbAbilities(token, abilities) : token));
    setLog((items) => ["Orb coletada: habilidades e efeitos passivos da unidade recuperados.", ...items].slice(0, 20));
  };
  // An HTTP action may finish after a visual movement collects an orb. Preserve
  // that recovery when applying the response, without changing other units.
  const preserveOrbRecoveries = (items: SkillToken[], requestRevision: number) => items.map((token) =>
    (unitOrbRevision.current.get(token.id) ?? 0) > requestRevision ? restoreOrbAbilities(token, abilities) : token);
  const activeAbility = abilities.find((ability) => ability.id === activeAbilityID);
  const baseStamina = config?.baseStamina ?? 5;
  const selected = tokens.find((token) => token.id === selectedID);
  const selectedWeapon = weapons.find((weapon) => weapon.id === selected?.weaponID);
  const selectedStamina = selected ? labStamina(selected, weapons, baseStamina) : null;
  const lookup = useMemo(() => new Map(board.map((tile) => [positionKey(tile), tile])), [board]);
  const occupied = useMemo(() => new Map(tokens.map((token) => [positionKey(token), token])), [tokens]);
  const movement = useMemo(() => {
    if (!selected || selected.life <= 0 || tool !== "select") return null;
    const units = tokens.map((token) => labMovementUnit(token, weapons, baseStamina));
    return movementOptions(board, labMovementUnit(selected, weapons, baseStamina), units);
  }, [selected, tokens, weapons, board, tool, baseStamina]);
  const attackTargets = new Set<string>();
  const attackArea = tool === "attack" && selected && selected.life > 0 && selectedWeapon && !weaponReloading(selectedWeapon, selected.turnCount)
    ? attackableArea(selected, selectedWeapon, board) : new Set<string>();
  const barricadeTargets = new Set<string>();
  if (tool === "attack" && selected) for (const target of tokens) {
    if (!labAttackProblem(selected, target, board, weapons)) attackTargets.add(positionKey(target));
  }
  if (tool === "attack" && selected) for (const tile of board) {
    if (tile.object==="barricade"&&!labAttackProblem(selected,undefined,board,weapons,tile)) barricadeTargets.add(positionKey(tile));
  }
  const selectable = new Set<string>();
  if (!pending) for (const tile of board) {
    const key = positionKey(tile);
    if ((tool === "paint" && !occupied.has(key))
      || (tool === "add" && isWalkable(tile) && !occupied.has(key))
      || (tool === "orb" && isWalkable(tile) && !occupied.has(key) && !orbs.some((orb) => positionKey(orb) === key))
      || (tool === "remove" && (occupied.has(key) || orbs.some((orb) => positionKey(orb) === key)))
      || (tool === "select" && (occupied.has(key) || movement?.paths.has(key)))
      || (tool === "skill" && !labAbilityProblem(selected, activeAbility, board, tokens, matchAbilityUses, tile))
      || (tool === "attack" && (attackTargets.has(key) || barricadeTargets.has(key)))) selectable.add(key);
  }
  const preview = tool === "select" && hovered ? movement?.paths.get(positionKey(hovered)) : undefined;
  const actors: MapActor[] = tokens.flatMap((token) => {
    const tile = lookup.get(positionKey(token));
    return tile ? [{ id: token.id, tile, path: token.path.map((point) => lookup.get(positionKey(point))).filter((point): point is MapTile => !!point),
      model: token.model, color: TEAMS[token.team], life: token.life, shield: token.shield, movementMode: token.movementMode,
      defeated: token.life <= 0, selected: token.id === selectedID && token.life > 0 }] : [];
  });
  const addLog = (message: string) => setLog((items) => [message, ...items].slice(0, 20));
  const chooseToken = (token: SkillToken) => { setSelectedID(token.id); setWeaponID(token.weaponID); setActiveAbilityID(null); if (tool === "skill") setTool("select"); setError(null); };

  const performAbility = async (id: string, target: MapTile) => {
    if (busy.current || !config || !selected) return;
    const ability = abilities.find((item) => item.id === id);
    const problem = labAbilityProblem(selected, ability, board, tokens, matchAbilityUses, target);
    if (problem) { setError(problem); return; }
    const requestOrbRevision = orbRevision.current;
    const controller = new AbortController(); activeRequest.current = controller;
    busy.current = true; setPending(true); setError(null);
    try {
      const result = await activateAbilityInLab(board,tokens,selected.id,id,target,matchAbilityUses,controller.signal);
      if (!controller.signal.aborted) {
        if (ability?.effect.some((effect) => effect.id === EffectID.areaDamageOrShieldBreak)) {
          setExplosions((items) => [...items, {id: crypto.randomUUID(), x: target.x, y: target.y}]);
        }
        setConfig((current) => current ? {...current,board:result.board} : current);
        setTokens(preserveOrbRecoveries(result.tokens, requestOrbRevision)); setMatchAbilityUses(result.matchAbilityUses); addLog(result.message);
        setActiveAbilityID(null); setTool("select"); setHovered(null);
      }
    } catch (cause) { if (!controller.signal.aborted) setError(errorText(cause)); }
    finally { if (activeRequest.current === controller) { busy.current = false; setPending(false); } }
  };
  const startAbility = (id: string) => {
    if (!selected) return;
    const ability = abilities.find((item) => item.id === id);
    const problem = labAbilityProblem(selected,ability,board,tokens,matchAbilityUses);
    if (problem) { setError(problem); return; }
    if (ability?.effect.some((effect) => effect.id === EffectID.grantShield)) {
      const tile = lookup.get(positionKey(selected)); if (tile) void performAbility(id,tile);
    } else { setActiveAbilityID(id); setTool("skill"); setHovered(null); setError(null); }
  };

  const onSelect = async (tile: MapTile) => {
    if (busy.current || !config) return;
    setError(null);
    const key = positionKey(tile);
    const token = tokens.find((item) => item.life > 0 && positionKey(item) === key) ?? occupied.get(key);
    if (tool === "orb") {
      if (!isWalkable(tile) || occupied.has(key) || orbs.some((orb) => positionKey(orb) === key)) { setError("Adicione a orb em um terreno acessível e livre."); return; }
      setOrbs((items) => [...items, {id:crypto.randomUUID(),x:tile.x,y:tile.y,triggered:false}]);
      return;
    }
    if (tool === "skill" && activeAbilityID) { await performAbility(activeAbilityID,tile); return; }
    if (tool === "attack") {
      if (!selected) { setError("Selecione o atacante na lista de tokens."); return; }
      const problem = labAttackProblem(selected, token, board, weapons,tile);
      if (problem) { setError(problem); return; }
      const requestOrbRevision = orbRevision.current;
      const controller = new AbortController();
      activeRequest.current = controller; busy.current = true; setPending(true);
      try {
        const result = await attackInLab(board, tokens, selected.id, token?.id, controller.signal,tile);
        if (!controller.signal.aborted) { setTokens(preserveOrbRecoveries(result.tokens, requestOrbRevision)); setConfig((current)=>current?{...current,board:result.board}:current); addLog(result.message); }
      } catch (cause) { if (!controller.signal.aborted) setError(errorText(cause)); }
      finally { if (activeRequest.current === controller) { busy.current = false; setPending(false); } }
      return;
    }
    if (tool === "remove") {
      if (token) { setTokens((items) => items.filter((item) => item.id !== token.id)); if (token.id === selectedID) setSelectedID(null); }
      else setOrbs((items) => items.filter((orb) => positionKey(orb) !== key));
      return;
    }
    if (tool === "add") {
      if (!isWalkable(tile) || occupied.has(key)) { setError("Adicione o token em um terreno acessível e livre."); return; }
      if (tokens.length >= 128) { setError("Limite de 128 tokens por cenário."); return; }
      const next: SkillToken = { id: crypto.randomUUID(), x: tile.x, y: tile.y, team, model, weaponID,
        life, maxLife: life, shield, maxShield: shield, abilityIDs: [...abilityIDs], staminaSpent: 0, path: [], turnCount: 1 };
      setTokens((items) => [...items, applyPassiveAbilities(next, abilities)]); setSelectedID(next.id);
      triggerOrb(tile, next.id);
      addLog(`Token da equipe ${team + 1} adicionado em (${tile.x}, ${tile.y}).`);
      return;
    }
    if (tool === "paint") {
      if (token) { setError("Remova ou mova o token antes de editar esse slot."); return; }
      const painted: MapTile = { x: tile.x, y: tile.y, terrain: terrain === "dirt" || terrain === "water" ? terrain : "grass",
        height, elevation: height, object: terrain === "house" || terrain === "tree" || terrain === "rock" ? terrain : undefined,
        roof: terrain === "house" ? "roofPurple" : undefined };
      setConfig({ ...config, board: board.map((cell) => positionKey(cell) === key ? painted : cell) });
      if (!isWalkable(painted)) setOrbs((items) => items.filter((orb) => positionKey(orb) !== key));
      return;
    }
    if (token && token.life > 0) { chooseToken(token); return; }
    const path = movement?.paths.get(key);
    const cost = movement?.costs.get(key);
    if (selected && path?.length && cost !== undefined) {
      setTokens((items) => items.map((item) => item.id === selected.id
        ? { ...item, x: tile.x, y: tile.y, movementMode: "", staminaSpent: (item.staminaSpent ?? 0) + cost, path: [{ x: item.x, y: item.y }, ...path] } : item));
    }
  };

  const equipWeapon = (id: string) => {
    setWeaponID(id);
    if (selected && tool !== "add") setTokens((items) => items.map((token) => token.id === selected.id ? { ...token, weaponID: id } : token));
    setError(null);
  };
  const equipAbility = (slot: number, id: string) => {
    setActiveAbilityID(null); if (tool === "skill") setTool("select");
    const current = tool === "add" ? abilityIDs : selected?.abilityIDs ?? [];
    if (id && (!abilities.some((ability) => ability.id === id) || current.some((item, index) => index !== slot && item === id))) return;
    const next = [current[0] ?? "", current[1] ?? ""];
    next[slot] = id;
    if (tool === "add") setAbilityIDs(next);
    else if (selected) setTokens((items) => items.map((token) => token.id === selected.id ? applyPassiveAbilities({ ...token, abilityIDs: next }, abilities) : token));
  };
  const recoverStamina = () => {
    if (!selected) return;
    setTokens((items) => items.map((token) => token.id === selected.id ? { ...token, staminaSpent: 0 } : token));
    setError(null); addLog("Estamina recuperada até o limite da arma equipada.");
  };
  const nextTokenTurn = () => {
    if (!selected || selected.life <= 0) return;
    setTokens((items) => items.map((token) => token.id === selected.id
      ? {...token, turnCount: Math.min(1000000, Math.max(1, token.turnCount ?? 1) + 1), staminaSpent: 0} : token));
    setActiveAbilityID(null); setTool("select"); setHovered(null); setError(null);
    addLog("Próximo turno do token: estamina recuperada e ciclo de recarga avançado.");
  };
  const restoreSelected = () => {
    if (!selected) return;
    if (tokens.some((token) => token.id !== selected.id && token.life > 0 && positionKey(token) === positionKey(selected))) {
      setError("Mova o token que ocupa esse slot antes de reviver a unidade."); return;
    }
    setTokens((items) => items.map((token) => token.id === selected.id ? { ...token, life: token.maxLife, shield: token.maxShield, staminaSpent: 0 } : token));
    // Remount stationary sprites at their authoritative slot when reviving.
    setVersion((value) => value + 1); setError(null); addLog("Vida, escudo e estamina restaurados.");
  };
  const reset = () => {
    setOrbs([]);
    setExplosions([]);
    collectedOrbs.current.clear(); unitOrbRevision.current.clear(); orbRevision.current = 0;
    if (!initial.current) return;
    const snapshot = structuredClone(initial.current);
    setConfig(snapshot.config); setTokens(snapshot.tokens); setSelectedID(snapshot.tokens[0]?.id ?? null);
    setActiveAbilityID(null); setMatchAbilityUses({});
    setTool("select"); setAbilityIDs([]); setHovered(null); setVersion((value) => value + 1); setError(null); setLog(["Cenário reiniciado."]);
  };

  return (
    <main className="skills-lab">
      <header className="lab-header">
        <div><Link to="/home">← Voltar ao início</Link><h1>Laboratório de skills</h1><p>Monte o cenário, equipe uma arma e experimente.</p></div>
        <span className="lab-badge">Modo criativo · ataques livres</span>
      </header>
      <form className="lab-scenario" onSubmit={(event) => { event.preventDefault(); void loadScenario(seed, layout); }}>
        <label>Cenário<select value={layout} disabled={pending} onChange={(event) => setLayout(event.target.value as LabLayout)}><option value="arena">Arena de testes</option><option value="procedural">Mapa procedural</option></select></label>
        <label>Seed<input value={seed} maxLength={100} disabled={pending} onChange={(event) => setSeed(event.target.value)} /></label>
        <button type="submit" disabled={pending}>{pending ? "Carregando…" : "Gerar cenário"}</button>
        <button type="button" disabled={pending || !config} onClick={reset}>Reiniciar cenário</button>
      </form>
      {error ? <div className="lab-error" role="alert">{error}</div> : null}
      <div className="lab-layout">
        <section className="lab-stage" aria-label="Cenário de testes">
          <div className="lab-tools" role="group" aria-label="Ferramentas do modo criativo">
            {TOOLS.map((item) => <button type="button" key={item.id} aria-pressed={tool === item.id} disabled={pending || !config || (item.id === "attack" && (!selected || selected.life <= 0))}
              onClick={() => { setTool(item.id); setHovered(null); setError(null); }}>{item.label}</button>)}
          </div>
          <p className="lab-hint" role="status">{pending ? "Preparando o teste…" : tool === "skill" ? `${activeAbility?.name}: clique no terreno alvo. Os destinos válidos estão destacados ao passar o mouse; a granada mostra a área da explosão.` : TOOLS.find((item) => item.id === tool)?.hint}</p>
          {config ? <ProceduralMap key={version} width={15} height={15} tiles={board} actors={actors} orbs={orbs} explosions={explosions} className="lab-map"
            onExplosionComplete={(id) => setExplosions((items) => items.filter((item) => item.id !== id))}
            onOrbComplete={(id) => setOrbs((items) => items.filter((orb) => orb.id !== id))}
            interaction={{ selectable, attackTargets, attackArea: tool === "attack" ? attackArea : undefined, onHover: setHovered, onActorTileEnter: triggerOrb, onSelect: (tile) => { void onSelect(tile); },
              area: tool === "skill" && hovered && selectable.has(positionKey(hovered)) ? abilityArea(activeAbility,hovered,board) : [],
              path: selected && preview ? [{ x: selected.x, y: selected.y }, ...preview] : [] }} />
            : <div className="lab-placeholder">{pending ? "Carregando o mapa e o catálogo de armas…" : "Use Gerar cenário para tentar novamente."}</div>}
        </section>
        <aside className="lab-sidebar">
          <fieldset disabled={pending || !config}>
            <legend>{tool === "add" ? "Novo token" : "Arma de teste"}</legend>
            <label>Arma<select value={tool === "add" ? weaponID : selected?.weaponID ?? weaponID} onChange={(event) => equipWeapon(event.target.value)}>
              {weapons.map((weapon) => <option key={weapon.id} value={weapon.id}>{weapon.icon} {weapon.name}</option>)}
            </select></label>
            {(() => { const weapon = tool === "add" ? weapons.find((item) => item.id === weaponID) : selectedWeapon;
              return weapon ? <div className="lab-weapon"><strong>Dano {weapon.damage} · Alcance {weapon.range}</strong><p>{weapon.description}</p>
                <ul className="lab-effects">{weapon.effect.map((effect) => <li key={effect.id} data-negative={effect.negative} title={effect.description}>
                  {effect.negative ? "−" : "+"} {effect.name}{effect.negative ? " (negativo)" : ""}
                </li>)}</ul>
                <p>Estamina: {tool === "add" ? effectiveStamina(baseStamina, [weapon]) : selectedStamina?.current ?? 0}/{effectiveStamina(baseStamina, [weapon])} · Base {baseStamina} − {weapon.staminaPenalty}</p>
                <small>Penalidade de estamina ativa. Penalidade de habilidades −{weapon.toolPenalty} ainda não aplicada.</small></div> : <p>Selecione um token ou adicione uma unidade.</p>; })()}
            {tool === "add" ? <div className="lab-input-grid">
              <label>Equipe<select value={team} onChange={(event) => setTeam(Number(event.target.value))}><option value={0}>1 · Azul</option><option value={1}>2 · Amarela</option></select></label>
              <label>Personagem<select value={model} onChange={(event) => setModel(Number(event.target.value) as CharacterModel)}><option value={0}>Modelo 1</option><option value={1}>Modelo 2</option><option value={2}>Modelo 3</option></select></label>
              <label>Vida<input type="number" min={1} max={99} value={life} onChange={(event) => setLife(Math.max(1, Math.min(99, Number(event.target.value))))} /></label>
              <label>Escudo<input type="number" min={0} max={99} value={shield} onChange={(event) => setShield(Math.max(0, Math.min(99, Number(event.target.value))))} /></label>
            </div> : null}
            {tool === "paint" ? <div className="lab-input-grid">
              <label>Terreno<select value={terrain} onChange={(event) => setTerrain(event.target.value as typeof terrain)}>
                <option value="grass">Grama</option><option value="dirt">Terra</option><option value="water">Água</option><option value="house">Casa</option><option value="tree">Árvore</option><option value="rock">Rocha</option>
              </select></label>
              <label>Elevação<select value={height} onChange={(event) => setHeight(Number(event.target.value))}>{[0, 1, 2, 3].map((level) => <option key={level} value={level}>Nível {level}</option>)}</select></label>
            </div> : null}
          </fieldset>
          <AbilitySelector abilities={abilities} value={tool === "add" ? abilityIDs : selected?.abilityIDs ?? []}
            onChange={equipAbility} addingToken={tool === "add"}
            onUse={startAbility} activeAbilityID={tool === "skill" ? activeAbilityID : null}
            activationProblem={(ability) => labAbilityProblem(selected,ability,board,tokens,matchAbilityUses)}
            disabled={pending || !config || (tool !== "add" && !selected)} />
          <section className="lab-card" aria-label="Tokens no cenário">
            <div className="lab-card-heading"><h2>Tokens <span>{tokens.length}</span></h2><button disabled={pending || !tokens.length} onClick={() => { setTokens([]); setSelectedID(null); setError(null); }}>Limpar</button></div>
            <div className="lab-token-list">{tokens.map((token, index) => <button key={token.id} disabled={pending} aria-pressed={selectedID === token.id}
              onClick={() => { chooseToken(token); if (tool === "add" || tool === "remove" || tool === "paint") setTool("select"); }}>
              <span className="lab-team-dot" style={{ background: TEAMS[token.team] }} />
              <span>Token {index + 1} · E{token.team + 1}<small>{token.life <= 0 ? "☠ Derrotado" : `Vida ${token.life}/${token.maxLife} · Escudo ${token.shield}/${token.maxShield}`} · ({token.x}, {token.y})</small><small>Estamina {labStamina(token, weapons, baseStamina).current}/{labStamina(token, weapons, baseStamina).max}</small></span>
            </button>)}</div>
            {selected ? <>
              <p>Turno do token: {Math.max(1, selected.turnCount ?? 1)} · {weaponReloading(selectedWeapon, selected.turnCount) ? "Recarregando" : "Pronto para atacar"}</p>
              <button className="lab-restore" disabled={pending || selected.life <= 0} onClick={nextTokenTurn}>Próximo turno do token</button>
              <button className="lab-restore" disabled={pending || selected.life <= 0 || selectedStamina?.current === selectedStamina?.max} onClick={recoverStamina}>Recuperar estamina</button>
              <button className="lab-restore" disabled={pending} onClick={restoreSelected}>Restaurar recursos do selecionado</button>
            </> : null}
          </section>
          <section className="lab-card"><h2>Registro de testes</h2><ol className="lab-log" aria-live="polite" aria-relevant="additions">{log.map((message, index) => <li key={`${log.length}-${index}`}>{message}</li>)}</ol></section>
        </aside>
      </div>
    </main>
  );
}
