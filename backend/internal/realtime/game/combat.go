package game

import (
	"errors"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

func (s *GameSession) Attack(id uuid.UUID, unitID string, target models.Position) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	p, err := s.actingUnit(id, unitID)
	if err != nil {
		return err
	}
	var attacker *models.UnitInstance
	for i := range p.Units {
		if p.Units[i].Unit.ID == unitID {
			attacker = &p.Units[i]
			break
		}
	}
	if attacker == nil || attacker.CurrentLife <= 0 {
		return errors.New("atacante inválido")
	}
	if attacker.HasAttacked {
		return errors.New("esta unidade já atacou neste turno")
	}
	if len(attacker.Unit.Weapons) == 0 {
		return errors.New("unidade sem arma")
	}
	weapon := attacker.Unit.Weapons[0]
	if weapon.Reloading(attacker.TurnCount) {
		return errors.New("arma em recarga neste turno; a unidade não pode atacar")
	}
	distance, valid := DirectAttackDistance(attacker.Position, target)
	if !valid {
		return errors.New("escolha uma das oito casas vizinhas ou um alvo na mesma linha ou coluna")
	}
	if distance > weapon.Range {
		return errors.New("alvo fora de alcance")
	}
	cells := make(map[models.Position]models.Cell, len(s.Board))
	for _, cell := range s.Board {
		cells[models.Position{X: cell.X, Y: cell.Y}] = cell
	}
	from, fromExists := cells[attacker.Position]
	to, toExists := cells[target]
	if (weapon.HasEffect(models.EffectSameLevelID) || weapon.HasEffect(models.EffectNoHigherTargetsID)) && (!fromExists || !toExists) {
		return errors.New("terreno do atacante ou alvo não encontrado")
	}
	if err := ValidateWeaponEffects(weapon, distance, from.Height, to.Height); err != nil {
		return err
	}
	if !DirectAttackLineClear(attacker.Position, target, cells) {
		return errors.New("uma estrutura ou o relevo bloqueia a linha de ataque")
	}
	if to.Object == models.ObjectBarricade {
		for i := range s.Board {
			if s.Board[i].X == target.X && s.Board[i].Y == target.Y {
				s.Board[i].Object = models.ObjectNone
				break
			}
		}
		attacker.HasAttacked = true
		s.changed(models.HistoryAction{Type: models.HistoryAttack, PlayerID: &id, UnitID: unitID, Target: &target})
		return nil
	}
	for owner, enemy := range s.Players {
		if owner == id {
			continue
		}
		for i := range enemy.Units {
			defender := &enemy.Units[i]
			if defender.Position != target || defender.CurrentLife <= 0 {
				continue
			}
			defender.CurrentLife, defender.CurrentShield = ApplyWeaponDamage(weapon, defender.CurrentLife, defender.CurrentShield)
			if defender.CurrentLife > 0 {
				if path := WeaponKnockback(weapon, attacker.Position, defender.Position, cells, s.occupied()); len(path) > 1 {
					defender.Path = path
					defender.MovementMode = ""
					defender.Position = path[len(path)-1]
					s.collectOrbs(defender, path[1:])
				}
			}
			attacker.HasAttacked = true
			alive := false
			for _, unit := range enemy.Units {
				if unit.CurrentLife > 0 {
					alive = true
					break
				}
			}
			if !alive {
				winner := id
				s.Winner = &winner
				s.Status = models.GameStatusFinished
				s.CurrentTurn = nil
				s.CurrentUnitID = ""
			}
			s.spawnMilestoneOrbs()
			s.changed(models.HistoryAction{Type: models.HistoryAttack, PlayerID: &id, UnitID: unitID, Target: &target})
			return nil
		}
	}
	return errors.New("alvo inimigo não encontrado")
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
