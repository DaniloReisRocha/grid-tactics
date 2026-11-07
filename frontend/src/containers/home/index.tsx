import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";

export const Home = () => {
  const navigate = useNavigate();
  const [accountID, setAccountID] = useState(() => sessionStorage.getItem("grid-tactics.account") ?? "");
  const [sessionID, setSessionID] = useState("");
  const [error, setError] = useState<string | null>(null);
  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const account = accountID.trim().toLowerCase();
    const session = sessionID.trim();
    if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/.test(account)
      || account === "00000000-0000-0000-0000-000000000000") {
      setError("Informe um UUID de conta válido ou gere uma conta de teste.");
      return;
    }
    if (!session || session.length > 100) { setError("Informe uma sessão de até 100 caracteres."); return; }
    sessionStorage.setItem("grid-tactics.account", account);
    navigate(`/grid-tactics/${encodeURIComponent(session)}/${encodeURIComponent(account)}`);
  };
  return (
    <main style={{ maxWidth: 440, margin: "auto", padding: 24 }}>
      <h1 style={{ fontSize: 30 }}>Grid Tactics</h1>
      <p>Use a mesma sessão e contas diferentes para jogar em duas abas.</p>
      <form onSubmit={submit} style={{ display: "grid", gap: 12 }}>
        <label htmlFor="account">Conta (UUID)</label>
        <input id="account" value={accountID} onChange={(event) => setAccountID(event.target.value)} required />
        <button type="button" onClick={() => { setAccountID(crypto.randomUUID()); setError(null); }}>Gerar conta de teste</button>
        <label htmlFor="session">Sessão</label>
        <input id="session" value={sessionID} onChange={(event) => setSessionID(event.target.value)} maxLength={100} required />
        {error ? <p role="alert">{error}</p> : null}
        <button type="submit">Entrar na partida</button>
      </form>
    </main>
  );
};
