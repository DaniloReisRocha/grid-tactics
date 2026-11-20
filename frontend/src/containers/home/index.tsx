import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { loadTeamOptions, type TeamOptions, type UnitLoadout } from "@/api/team";
import { defaultTeam, loadTeamDraft, saveMatchTeam, validateTeam } from "@/internal/game/team";
import TeamUnitCard from "./TeamUnitCard";
import "./styles.css";

export const Home = () => {
  const navigate = useNavigate();
  const [accountID, setAccountID] = useState(() => sessionStorage.getItem("grid-tactics.account") ?? "");
  const [sessionID, setSessionID] = useState("");
  const [team, setTeam] = useState<UnitLoadout[]>(loadTeamDraft);
  const [options, setOptions] = useState<TeamOptions | null>(null);
  const [loading, setLoading] = useState(true);
  const [catalogError, setCatalogError] = useState<string | null>(null);
  const [retry, setRetry] = useState(0);
  const [error, setError] = useState<string | null>(null);
  useEffect(() => {
    const controller = new AbortController(); setLoading(true); setCatalogError(null);
    loadTeamOptions(controller.signal).then((data) => { if (!controller.signal.aborted) setOptions(data); })
      .catch((cause: unknown) => { if (!controller.signal.aborted) setCatalogError(cause instanceof Error ? cause.message : "Não foi possível carregar os equipamentos."); })
      .finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [retry]);
  useEffect(() => { try { sessionStorage.setItem("grid-tactics.team.draft", JSON.stringify(team)); } catch { /* Submit reports unavailable storage. */ } }, [team]);
  const teamProblem = options ? validateTeam(team, options) : null;
  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const account = accountID.trim().toLowerCase();
    const session = sessionID.trim();
    if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/.test(account)
      || account === "00000000-0000-0000-0000-000000000000") {
      setError("Informe um UUID de conta válido ou gere uma conta de teste.");
      return;
    }
    if (!session || session.length > 100) { setError("Informe um código de sala de até 100 caracteres."); return; }
    if (!options || teamProblem || loading) { setError(teamProblem ?? "Aguarde o catálogo de equipamentos."); return; }
    try { sessionStorage.setItem("grid-tactics.account", account); saveMatchTeam(session, account, team); }
    catch { setError("Não foi possível salvar a equipe nesta aba. Habilite o armazenamento do navegador para entrar."); return; }
    navigate(`/grid-tactics/${encodeURIComponent(session)}/${encodeURIComponent(account)}`);
  };
  return (
    <main className="team-home">
      <header className="team-home-header"><div><span className="team-eyebrow">GRID TACTICS</span><h1>Prepare sua equipe</h1><p>Três unidades. Escolha o equipamento de cada uma e entre em combate.</p></div><Link to="/skills-lab">Laboratório de skills ↗</Link></header>
      <form className="team-home-layout" onSubmit={submit}>
        <section className="team-builder" aria-label="Montagem de equipe">
          <div className="team-section-heading"><div><h2>Sua formação</h2><p>Vida 3 · Estamina base {options?.baseStamina ?? 5} · Escudo adquirido por habilidade</p></div><button type="button" disabled={loading} onClick={() => { setTeam(defaultTeam()); setError(null); }}>Restaurar equipe</button></div>
          {catalogError ? <div className="team-alert" role="alert">{catalogError}<button type="button" onClick={() => setRetry((value) => value + 1)}>Tentar novamente</button></div> : null}
          {loading ? <p className="team-loading" role="status">Carregando armas e habilidades…</p> : null}
          {options ? <fieldset className="team-cards" disabled={loading}><legend className="team-sr-only">Equipamentos das três unidades</legend>
            {team.map((unit, index) => <TeamUnitCard key={index} index={index} value={unit} options={options} onChange={(next) => { setTeam((current) => current.map((item, i) => i === index ? next : item)); setError(null); }} />)}
          </fieldset> : null}
          {teamProblem ? <p className="team-alert" role="alert">{teamProblem}</p> : null}
        </section>
        <aside className="team-join">
          <span className="team-eyebrow">PRONTO PARA JOGAR?</span><h2>Entrar na partida</h2>
        <label htmlFor="account">Conta (UUID)</label>
        <input id="account" value={accountID} onChange={(event) => setAccountID(event.target.value)} required />
        <button className="team-account-button" type="button" onClick={() => { setAccountID(crypto.randomUUID()); setError(null); }}>Gerar conta de teste</button>
        <label htmlFor="session">Código da sala</label>
        <input id="session" value={sessionID} onChange={(event) => setSessionID(event.target.value)} placeholder="Digite ou gere um código" maxLength={100} required />
        <button className="team-account-button" type="button" onClick={() => { setSessionID(crypto.randomUUID().slice(0, 8).toUpperCase()); setError(null); }}>Gerar código da sala</button>
          <p>Para jogar em duas abas, use o mesmo código da sala e contas diferentes.</p>
          <dl className="team-join-summary"><div><dt>Unidades</dt><dd>{team.length}/3</dd></div><div><dt>Armas</dt><dd>{team.filter((unit) => unit.weaponID).length}</dd></div><div><dt>Habilidades equipadas</dt><dd>{team.reduce((sum, unit) => sum + unit.abilityIDs.filter(Boolean).length, 0)}</dd></div></dl>
          {error ? <p className="team-alert" role="alert">{error}</p> : null}
          <button className="team-enter" type="submit" disabled={loading || !options || !!catalogError || !!teamProblem}>Entrar com esta equipe</button>
          <p className="team-join-note">A equipe é confirmada ao entrar. Ao reconectar, a formação já registrada na partida é mantida.</p>
        </aside>
      </form>
    </main>
  );
};
