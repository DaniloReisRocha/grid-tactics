import { useGameSocket } from "@/store/GameSocket";

export const TurnControl = () => {
  const { currentTurn, turnCounter, players, accountID, endTurn, pending, status, connection } = useGameSocket();
  const player = currentTurn ? players[currentTurn] : undefined;
  const yourTurn = currentTurn === accountID;
  return (
    <section>
      <p>Turno {turnCounter}: {player?.name ?? "—"}</p>
      <p>{yourTurn ? "É a sua vez." : "Aguarde sua vez."}</p>
      <button type="button" onClick={endTurn}
        disabled={!yourTurn || pending || status !== "playing" || connection !== "connected"}>
        Finalizar turno
      </button>
    </section>
  );
};
