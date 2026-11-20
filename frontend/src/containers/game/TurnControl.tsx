import { GameStatus, ConnectionStatus } from "@/api/status";
import { useGameSocket } from "@/store/GameSocket";
import { useTurnSeconds } from "./useTurnSeconds";

export const TurnControl = () => {
  const { currentTurn, currentUnitID, roundCounter, roundTurn, roundOrder, players,
    accountID, endTurn, pending, status, connection, setUnit } = useGameSocket();
  const seconds = useTurnSeconds();
  const player = currentTurn ? players[currentTurn] : undefined;
  const active = player?.units.find((unit) => unit.unit.id === currentUnitID);
  const yourTurn = currentTurn === accountID;
  return (
    <section style={{ flexBasis: "100%" }}>
      <p>Round {roundCounter} · Turno {roundTurn + 1}/{roundOrder.length}</p>
      <p>{player?.name ?? "—"} · {active?.unit.icon} Unidade {player ? player.units.findIndex((unit) => unit.unit.id === currentUnitID) + 1 : "—"}</p>
      <p>{yourTurn ? "É a vez desta unidade sua." : "Aguarde sua vez."}</p>
      <button type="button" onClick={endTurn}
        disabled={!yourTurn || pending || status !== GameStatus.playing || connection !== ConnectionStatus.connected || seconds === 0}>
        Finalizar turno da unidade
      </button>
      <ol aria-label="Sequência do round" style={{ display: "flex", flexWrap: "wrap", gap: 8, padding: 0, listStyle: "none" }}>
        {roundOrder.map((turn, index) => {
          const owner = players[turn.playerID];
          const unitIndex = owner?.units.findIndex((unit) => unit.unit.id === turn.unitID) ?? -1;
          const unit = owner?.units[unitIndex];
          if (!unit) return null;
          const current = index === roundTurn && Boolean(currentUnitID);
          return <li key={`${turn.playerID}:${turn.unitID}`}>
            <button type="button" onClick={() => setUnit(turn.unitID)} aria-current={current ? "step" : undefined}
              style={{ opacity: index < roundTurn || unit.currentLife <= 0 ? 0.4 : 1, border: current ? "2px solid #b1de73" : undefined }}>
              {index + 1}. {owner.name} · {unit.currentLife > 0 ? unit.unit.icon : "☠"} {unitIndex + 1}{current ? " · Agora" : ""}
            </button>
          </li>;
        })}
      </ol>
    </section>
  );
};
