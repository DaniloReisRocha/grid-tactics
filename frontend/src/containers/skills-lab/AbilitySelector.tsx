import { useState } from "react";
import type { Ability } from "@/api/models";

interface Props {
  abilities: readonly Ability[];
  value: readonly string[];
  onChange: (slot: number, id: string) => void;
  disabled: boolean;
  addingToken: boolean;
  onUse: (id: string) => void;
  activationProblem: (ability: Ability) => string | null;
  activeAbilityID?: string | null;
}

export default function AbilitySelector({ abilities, value, onChange, disabled, addingToken, onUse, activationProblem, activeAbilityID }: Props) {
  const [inspectedID, setInspectedID] = useState("");
  const ability = abilities.find((item) => item.id === inspectedID && value.includes(item.id))
    ?? abilities.find((item) => value.includes(item.id));
  return (
    <fieldset disabled={disabled || !abilities.length}>
      <legend>Skills · {value.filter(Boolean).length}/2</legend>
      <p className="lab-skill-hint">{addingToken ? "Habilidades do novo token" : "Habilidades do token selecionado"}</p>
      <div className="lab-skill-slots">
        {[0, 1].map((slot) => (
          <label key={slot}>Habilidade {slot + 1}
            <select value={value[slot] ?? ""} onChange={(event) => {
              onChange(slot, event.target.value); setInspectedID(event.target.value);
            }}>
              <option value="">Nenhuma</option>
              {abilities.map((item) => <option key={item.id} value={item.id}
                disabled={value.some((id, index) => index !== slot && id === item.id)}>
                {item.icon} {item.name}
              </option>)}
            </select>
          </label>
        ))}
      </div>
      {!addingToken ? <div className="lab-skill-actions">{value.filter(Boolean).map((id) => {
        const item = abilities.find((entry) => entry.id === id);
        if (!item) return null;
        const problem = activationProblem(item);
        return <div key={id}>
          <button type="button" disabled={disabled || !!problem} title={problem ?? item.description}
            aria-pressed={activeAbilityID === id} onClick={() => { setInspectedID(id); onUse(id); }}>Usar {item.name}</button>
          {problem ? <small>{problem}</small> : null}
        </div>;
      })}</div> : null}
      {!abilities.length ? <p className="lab-skill-hint">Catálogo indisponível. Atualize e reinicie o backend.</p> : null}
      {ability ? <div className="lab-weapon">
        <strong>{ability.icon} {ability.name}</strong>
        <p>{ability.description}</p>
        <dl className="lab-skill-stats">
          <div><dt>Uso</dt><dd>{ability.uses === "single" ? "Único" : "Ilimitado"}</dd></div>
          <div><dt>Alcance</dt><dd>{ability.range === 0 ? "Própria unidade" : `${ability.range} terreno(s)`}</dd></div>
          {ability.shield > 0 ? <div><dt>Escudo</dt><dd>{ability.shield}</dd></div> : null}
          {ability.damage > 0 ? <div><dt>Dano</dt><dd>{ability.damage}</dd></div> : null}
          {ability.areaSize > 0 ? <div><dt>Área</dt><dd>{ability.areaSize} × {ability.areaSize}</dd></div> : null}
          {ability.maxUsesPerMatch > 0 ? <div><dt>Limite por partida</dt><dd>{ability.maxUsesPerMatch} uso(s)</dd></div> : null}
          {ability.maxActivePerMatch > 0 ? <div><dt>Limite de estruturas ativas</dt><dd>{ability.maxActivePerMatch}</dd></div> : null}
        </dl>
        <ul className="lab-effects">{ability.effect.map((effect) => <li key={effect.id} data-negative={effect.negative} title={effect.description}>
          {effect.negative ? "−" : "+"} {effect.name}{effect.negative ? " (negativo)" : ""}
        </li>)}</ul>
      </div> : <p className="lab-skill-hint">Selecione uma habilidade para ver os detalhes.</p>}
    </fieldset>
  );
}
