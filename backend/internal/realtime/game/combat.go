package game

import (
	"errors"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
)

func (s *GameSession) Attack(id uuid.UUID, unitID string, target models.Position) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	p, err := s.actingPlayer(id)
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
	dx, dy := abs(attacker.X-target.X), abs(attacker.Y-target.Y)
	if max(dx, dy) > weapon.Range {
		return errors.New("alvo fora de alcance")
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
			damage := max(0, weapon.Damage)
			absorbed := min(defender.CurrentShield, damage)
			defender.CurrentShield -= absorbed
			defender.CurrentLife = max(0, defender.CurrentLife-(damage-absorbed))
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
				s.Status = "finished"
				s.CurrentTurn = nil
			}
			s.changed()
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
