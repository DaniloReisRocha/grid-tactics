import type { TeamOptions, UnitLoadout } from "@/api/team";
import { abilitySlotCount } from "@/internal/game/team";
import { effectiveStamina } from "@/internal/game/stamina";

interface Props { index: number; value: UnitLoadout; options: TeamOptions; onChange: (value: UnitLoadout) => void }
const icons = ["🦸‍♂️", "🦸‍♀️", "🧙‍♂️"];

export default function TeamUnitCard({ index, value, options, onChange }: Props) {
  const weapon = options.weapons.find((item) => item.id === value.weaponID);
  const slots = abilitySlotCount(weapon?.toolPenalty ?? 0, options.abilitySlots);
  return <article className="team-unit" aria-label={`Equipamentos da unidade ${index + 1}`}>
    <header><span className="team-unit-avatar" aria-hidden="true">{icons[index]}</span><div><small>POSIÇÃO {index + 1}</small><h3>Unidade {index + 1}</h3></div></header>
    <dl className="team-resources">
      <div><dt>Vida</dt><dd>{options.baseLife}</dd></div><div><dt>Escudo</dt><dd>0/{options.shieldCapacity}</dd></div>
      <div><dt>Estamina</dt><dd>{effectiveStamina(options.baseStamina, weapon ? [weapon] : [])}/{options.baseStamina}</dd></div>
    </dl>
    <label htmlFor={`unit-${index}-weapon`}>Arma · {options.itemSlots} slot de item</label>
    <select id={`unit-${index}-weapon`} value={value.weaponID} onChange={(event) => {
      const nextWeapon = options.weapons.find((item) => item.id === event.target.value);
      const nextSlots = abilitySlotCount(nextWeapon?.toolPenalty ?? 0, options.abilitySlots);
      onChange({ weaponID: event.target.value, abilityIDs: value.abilityIDs.filter(Boolean).slice(0, nextSlots) });
    }}>{options.weapons.map((item) => <option key={item.id} value={item.id}>{item.icon} {item.name}</option>)}</select>
    {weapon ? <div className="team-weapon-info"><strong>Dano {weapon.damage} · Alcance {weapon.range}</strong><p>{weapon.description}</p>
      <ul>{weapon.effect.map((effect) => <li key={effect.id} data-negative={effect.negative} title={effect.description}>{effect.negative ? "−" : "+"} {effect.name}</li>)}</ul>
    </div> : null}
    <div className="team-skills-title"><strong>Itens / habilidades</strong><span>{value.abilityIDs.filter(Boolean).length}/{slots}</span></div>
    {Array.from({ length: options.abilitySlots }, (_, slot) => <div className="team-skill" key={slot}>
      <label htmlFor={`unit-${index}-ability-${slot}`}>Habilidade {slot + 1}{slot >= slots ? " · bloqueada pela arma" : ""}</label>
      <select id={`unit-${index}-ability-${slot}`} disabled={slot >= slots} value={value.abilityIDs[slot] ?? ""} onChange={(event) => {
        const next = [...value.abilityIDs]; next[slot] = event.target.value; onChange({ ...value, abilityIDs: next });
      }}><option value="">Nenhuma</option>{options.abilities.map((item) => <option key={item.id} value={item.id} disabled={value.abilityIDs.some((id, itemSlot) => itemSlot !== slot && id === item.id)}>{item.icon} {item.name}</option>)}</select>
      {slot < slots && value.abilityIDs[slot] ? <p>{options.abilities.find((item) => item.id === value.abilityIDs[slot])?.description}</p> : null}
    </div>)}
    <p className="team-unit-note">{weapon?.toolPenalty ? `Esta arma reduz os slots de habilidades de ${options.abilitySlots} para ${slots}. Trocar de arma ajusta as habilidades equipadas.` : "Escolha até duas habilidades diferentes para esta unidade."}</p>
  </article>;
}
