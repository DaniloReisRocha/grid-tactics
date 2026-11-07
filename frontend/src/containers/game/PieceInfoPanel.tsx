import { useGameSocket } from "@/store/GameSocket";
import { canUnitAttack } from "@/internal/game/combat";
import "@/styles/game/PieceInfoPanel.css";

export const PieceInfoPanel = () => {
  const {
    players,
    unitID,
    setUnit,
    accountID,
    attackMode,
    setAttackMode,
    currentTurn,
    status,
    pending,
    connection,
  } = useGameSocket();
  const units = Object.values(players).flatMap((player) => player.units);
  const selected = units.find((unit) => unit.unit.id === unitID);
  const canAct =
    selected?.ownerID === accountID &&
    currentTurn === accountID &&
    status === "playing" &&
    !pending &&
    connection === "connected";
  const weapon = selected?.unit.weapons[0];

  return (
    <aside className="unit-info-panel">
      {selected ? (
        <div className="selected-unit">
          <div className="icon">{selected.unit.icon}</div>
          <div>{players[selected.ownerID]?.name}</div>
          <div>
            Vida: {selected.currentLife}/{selected.unit.lp}
          </div>
          <div>
            Escudo: {selected.currentShield}/{selected.unit.shield}
          </div>
          <div>
            Stamina: {selected.currentStamina}/{selected.unit.stamina}
          </div>
          <div>
            Ataques disponíveis neste turno: {selected.hasAttacked ? 0 : 1}
          </div>
          {weapon ? (
            <div>
              {weapon.icon} {weapon.name} · Alcance {weapon.range} · Dano{" "}
              {weapon.damage}
            </div>
          ) : null}
          <button
            type="button"
            className={`attack-btn ${attackMode ? "pressed" : ""}`}
            disabled={!canAct || !canUnitAttack(selected)}
            onClick={() => setAttackMode(!attackMode)}
          >
            {selected.hasAttacked
              ? "Ataque utilizado"
              : attackMode
                ? "Cancelar ataque"
                : "Atacar"}
          </button>
        </div>
      ) : (
        <p>Selecione uma unidade no tabuleiro.</p>
      )}
      {Object.values(players).map((player) => (
        <section key={player.accoID}>
          <h3>
            {player.name}
            {player.accoID === accountID ? " (você)" : ""}
          </h3>
          {player.units.map((unit) => (
            <button
              type="button"
              key={unit.unit.id}
              disabled={unit.currentLife <= 0}
              className={`other-unit ${unit.unit.id === unitID ? "selected" : ""}`}
              onClick={() => setUnit(unit.unit.id)}
            >
              {unit.currentLife > 0 ? unit.unit.icon : "☠️"} · Vida{" "}
              {unit.currentLife} · Stamina {unit.currentStamina}
            </button>
          ))}
        </section>
      ))}
    </aside>
  );
};
