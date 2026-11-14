import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import ProceduralMap from "@/components/ProceduralMap";
import { isWalkable } from "@/components/ProceduralMap/characterMovement";
import type { MapTile } from "@/components/ProceduralMap/generator";
import type { CharacterModel, MapActor } from "@/components/ProceduralMap/types";
import { attackInLab, loadLab, type LabConfig, type LabLayout, type SkillToken } from "@/api/skillsLab";
import { movementOptions, positionKey } from "@/internal/game/pathfinding";
import { labAttackProblem, labMovementUnit, starterTokens } from "@/internal/game/skillsLab";
import "./styles.css";

type Tool = "select" | "add" | "attack" | "remove" | "paint";
const TOOLS: { id: Tool; label: string; hint: string }[] = [
  { id: "select", label: "Selecionar / mover", hint: "Selecione um token. Passe o mouse para ver a rota e clique em um slot para mover." },
  { id: "add", label: "+ Token", hint: "Escolha equipe, personagem e arma. Clique em um slot livre para adicionar." },
  { id: "attack", label: "Testar ataque", hint: "Clique em um alvo da outra equipe. Os alvos válidos estão marcados em vermelho." },
  { id: "remove", label: "Remover token", hint: "Clique em um token ou caveira para remover do cenário." },
  { id: "paint", label: "Editar terreno", hint: "Escolha o terreno e a altura. Clique em um slot sem tokens para pintar." },
];
const TEAMS = ["#68d2ff", "#ffcc66"];
const errorText = (error: unknown) => error instanceof Error ? error.message : "Não foi possível executar a ação.";

export default function SkillsLab() {
  const [config, setConfig] = useState<LabConfig | null>(null);
  const [tokens, setTokens] = useState<SkillToken[]>([]);
  const [selectedID, setSelectedID] = useState<string | null>(null);
  const [tool, setTool] = useState<Tool>("select");
  const [hovered, setHovered] = useState<MapTile | null>(null);
  const [seed, setSeed] = useState("skills-lab");
  const [layout, setLayout] = useState<LabLayout>("arena");
  const [team, setTeam] = useState(0);
  const [model, setModel] = useState<CharacterModel>(0);
  const [weaponID, setWeaponID] = useState("pistol");
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
      setWeaponID(data.weapons[0]?.id ?? "pistol"); setTool("select");
      setHovered(null); setVersion((value) => value + 1); setLog(["Cenário pronto. Ataques livres e movimento sem limite de estamina."]);
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
  const selected = tokens.find((token) => token.id === selectedID);
  const selectedWeapon = weapons.find((weapon) => weapon.id === selected?.weaponID);
  const lookup = useMemo(() => new Map(board.map((tile) => [positionKey(tile), tile])), [board]);
  const occupied = useMemo(() => new Map(tokens.map((token) => [positionKey(token), token])), [tokens]);
  const movement = useMemo(() => {
    if (!selected || selected.life <= 0 || tool !== "select") return null;
    const units = tokens.map((token) => labMovementUnit(token, weapons));
    return movementOptions(board, labMovementUnit(selected, weapons), units);
  }, [selected, tokens, weapons, board, tool]);
  const attackTargets = new Set<string>();
  if (tool === "attack" && selected) for (const target of tokens) {
    if (!labAttackProblem(selected, target, board, weapons)) attackTargets.add(positionKey(target));
  }
  const selectable = new Set<string>();
  if (!pending) for (const tile of board) {
    const key = positionKey(tile);
    if ((tool === "paint" && !occupied.has(key))
      || (tool === "add" && isWalkable(tile) && !occupied.has(key))
      || (tool === "remove" && occupied.has(key))
      || (tool === "select" && (occupied.has(key) || movement?.paths.has(key)))
      || (tool === "attack" && attackTargets.has(key))) selectable.add(key);
  }
  const preview = tool === "select" && hovered ? movement?.paths.get(positionKey(hovered)) : undefined;
  const actors: MapActor[] = tokens.flatMap((token) => {
    const tile = lookup.get(positionKey(token));
    return tile ? [{ id: token.id, tile, path: token.path.map((point) => lookup.get(positionKey(point))).filter((point): point is MapTile => !!point),
      model: token.model, color: TEAMS[token.team], life: token.life, shield: token.shield,
      defeated: token.life <= 0, selected: token.id === selectedID && token.life > 0 }] : [];
  });
  const addLog = (message: string) => setLog((items) => [message, ...items].slice(0, 20));
  const chooseToken = (token: SkillToken) => { setSelectedID(token.id); setWeaponID(token.weaponID); setError(null); };

  const onSelect = async (tile: MapTile) => {
    if (busy.current || !config) return;
    setError(null);
    const key = positionKey(tile);
    const token = tokens.find((item) => item.life > 0 && positionKey(item) === key) ?? occupied.get(key);
    if (tool === "attack") {
      if (!selected) { setError("Selecione o atacante na lista de tokens."); return; }
      const problem = labAttackProblem(selected, token, board, weapons);
      if (problem || !token) { setError(problem); return; }
      const controller = new AbortController();
      activeRequest.current = controller; busy.current = true; setPending(true);
      try {
        const result = await attackInLab(board, tokens, selected.id, token.id, controller.signal);
        if (!controller.signal.aborted) { setTokens(result.tokens); addLog(result.message); }
      } catch (cause) { if (!controller.signal.aborted) setError(errorText(cause)); }
      finally { if (activeRequest.current === controller) { busy.current = false; setPending(false); } }
      return;
    }
    if (tool === "remove") {
      if (token) { setTokens((items) => items.filter((item) => item.id !== token.id)); if (token.id === selectedID) setSelectedID(null); }
      return;
    }
    if (tool === "add") {
      if (!isWalkable(tile) || occupied.has(key)) { setError("Adicione o token em um terreno acessível e livre."); return; }
      if (tokens.length >= 128) { setError("Limite de 128 tokens por cenário."); return; }
      const next: SkillToken = { id: crypto.randomUUID(), x: tile.x, y: tile.y, team, model, weaponID,
        life, maxLife: life, shield, maxShield: shield, path: [] };
      setTokens((items) => [...items, next]); setSelectedID(next.id);
      addLog(`Token da equipe ${team + 1} adicionado em (${tile.x}, ${tile.y}).`);
      return;
    }
    if (tool === "paint") {
      if (token) { setError("Remova ou mova o token antes de editar esse slot."); return; }
      const painted: MapTile = { x: tile.x, y: tile.y, terrain: terrain === "dirt" || terrain === "water" ? terrain : "grass",
        height, elevation: height, object: terrain === "house" || terrain === "tree" || terrain === "rock" ? terrain : undefined,
        roof: terrain === "house" ? "roofPurple" : undefined };
      setConfig({ ...config, board: board.map((cell) => positionKey(cell) === key ? painted : cell) });
      return;
    }
    if (token && token.life > 0) { chooseToken(token); return; }
    const path = movement?.paths.get(key);
    if (selected && path?.length) {
      setTokens((items) => items.map((item) => item.id === selected.id
        ? { ...item, x: tile.x, y: tile.y, path: [{ x: item.x, y: item.y }, ...path] } : item));
    }
  };

  const equipWeapon = (id: string) => {
    setWeaponID(id);
    if (selected && tool !== "add") setTokens((items) => items.map((token) => token.id === selected.id ? { ...token, weaponID: id } : token));
    setError(null);
  };
  const restoreSelected = () => {
    if (!selected) return;
    if (tokens.some((token) => token.id !== selected.id && token.life > 0 && positionKey(token) === positionKey(selected))) {
      setError("Mova o token que ocupa esse slot antes de reviver a unidade."); return;
    }
    setTokens((items) => items.map((token) => token.id === selected.id ? { ...token, life: token.maxLife, shield: token.maxShield } : token));
    // Remount stationary sprites at their authoritative slot when reviving.
    setVersion((value) => value + 1); setError(null); addLog("Vida e escudo restaurados.");
  };
  const reset = () => {
    if (!initial.current) return;
    const snapshot = structuredClone(initial.current);
    setConfig(snapshot.config); setTokens(snapshot.tokens); setSelectedID(snapshot.tokens[0]?.id ?? null);
    setTool("select"); setHovered(null); setVersion((value) => value + 1); setError(null); setLog(["Cenário reiniciado."]);
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
          <p className="lab-hint" role="status">{pending ? "Preparando o teste…" : TOOLS.find((item) => item.id === tool)?.hint}</p>
          {config ? <ProceduralMap key={version} width={15} height={15} tiles={board} actors={actors} className="lab-map"
            interaction={{ selectable, attackTargets, onHover: setHovered, onSelect: (tile) => { void onSelect(tile); },
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
                <small>Penalidades: estamina −{weapon.staminaPenalty} · habilidades −{weapon.toolPenalty}. Desativadas neste modo.</small></div> : <p>Selecione um token ou adicione uma unidade.</p>; })()}
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
          <section className="lab-card" aria-label="Tokens no cenário">
            <div className="lab-card-heading"><h2>Tokens <span>{tokens.length}</span></h2><button disabled={pending || !tokens.length} onClick={() => { setTokens([]); setSelectedID(null); setError(null); }}>Limpar</button></div>
            <div className="lab-token-list">{tokens.map((token, index) => <button key={token.id} disabled={pending} aria-pressed={selectedID === token.id}
              onClick={() => { chooseToken(token); if (tool === "add" || tool === "remove" || tool === "paint") setTool("select"); }}>
              <span className="lab-team-dot" style={{ background: TEAMS[token.team] }} />
              <span>Token {index + 1} · E{token.team + 1}<small>{token.life <= 0 ? "☠ Derrotado" : `Vida ${token.life}/${token.maxLife} · Escudo ${token.shield}/${token.maxShield}`} · ({token.x}, {token.y})</small></span>
            </button>)}</div>
            {selected ? <button className="lab-restore" disabled={pending} onClick={restoreSelected}>Restaurar vida e escudo do selecionado</button> : null}
          </section>
          <section className="lab-card"><h2>Registro de testes</h2><ol className="lab-log" aria-live="polite" aria-relevant="additions">{log.map((message, index) => <li key={`${log.length}-${index}`}>{message}</li>)}</ol></section>
        </aside>
      </div>
    </main>
  );
}
