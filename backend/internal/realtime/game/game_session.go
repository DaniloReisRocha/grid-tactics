package game

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DaniloReisRocha/grid-tactics/backend/internal/realtime"
	"github.com/DaniloReisRocha/grid-tactics/backend/internal/utils"
	"github.com/DaniloReisRocha/grid-tactics/backend/internal/utils/array"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

const (
	gameStateUpdate = "gameStateUpdate"
)

type PlayerState struct {
	models.Player
	Connection *realtime.WSConnection
}

type GameSession struct {
	ID             string
	maximumPlayers int
	Players        map[uuid.UUID]*PlayerState
	Board          []models.Cell
	TurnOrder      []uuid.UUID
	CurrentTurn    *uuid.UUID
	TurnCounter    int
	StartedAt      time.Time
	LastAction     time.Time
	lock           *sync.RWMutex
}

func NewGameSession(id string, maximumPlayers int) *GameSession {
	return &GameSession{
		ID:             id,
		maximumPlayers: maximumPlayers,
		Players:        make(map[uuid.UUID]*PlayerState),
		Board:          generateBoard(id, 0.3),
		TurnOrder:      []uuid.UUID{},
		StartedAt:      time.Now(),
		LastAction:     time.Now(),
		lock:           &sync.RWMutex{},
	}
}

func (s *GameSession) BroadcastState() error {

	players := map[uuid.UUID]*models.Player{}
	for key, p := range s.Players {
		players[key] = &p.Player
	}
	payload := models.GameStatePayload{
		ID:          s.ID,
		Board:       s.Board,
		Players:     players,
		TurnOrder:   s.TurnOrder,
		CurrentTurn: s.CurrentTurn,
		TurnCounter: s.TurnCounter,
	}

	e := map[uuid.UUID]error{}
	for key, p := range s.Players {
		err := p.Connection.Send(realtime.ActionSent{Type: gameStateUpdate, Payload: payload})
		if err != nil {
			e[key] = err
		}
	}
	return formatUUIDErrors(e)
}

func (gs *GameSession) AddPlayer(conn *realtime.WSConnection, accoID uuid.UUID, units []models.Unit) error {
	gs.lock.Lock()
	defer gs.lock.Unlock()

	if len(gs.Players) >= gs.maximumPlayers {
		return errors.New("closed match")
	}

	if _, exists := gs.Players[accoID]; exists {
		return nil
	}

	used := map[string]bool{}
	for _, player := range gs.Players {
		if player != nil {
			u := buildUsedFromPiecePoints(player.Units)
			used = utils.MergeMaps(used, u)
		}
	}

	gs.Players[accoID] = &PlayerState{
		Player: models.Player{
			AccoID: accoID,
			Units:  generateInitialPieces(gs.ID, units, used),
		},
		Connection: conn,
	}

	gs.TurnOrder = append(gs.TurnOrder, accoID)
	return nil
}

func (gs *GameSession) MovePiece(accoID uuid.UUID, unitID string, to models.Position) error {
	gs.lock.Lock()
	defer gs.lock.Unlock()

	// if gs.CurrentTurn == nil {
	// 	return errors.New("turn not found")
	// }
	// if *gs.CurrentTurn != accoID {
	// 	return errors.New("invalid turn")
	// }
	player, ok := gs.Players[accoID]
	if !ok {
		return errors.New("player not found")
	}
	unitIndex, ok := array.FindIndex(player.Units, func(p models.UnitInstance) bool {
		return p.Unit.ID == unitID
	})
	if !ok {
		return errors.New("unit not found")
	}
	unit := player.Units[unitIndex]
	path := findPathToTarget(gs.Board, unit.X, unit.Y, to.X, to.Y)
	if path == nil {
		return errors.New("invalid path")
	}
	totalCost := 0
	for _, step := range path {
		cell, ok := findCell(gs.Board, step.X, step.Y)
		if !ok {
			return errors.New("invalid cell in path")
		}
		totalCost += 1 + int(cell.Elevation)
	}

	if totalCost > unit.CurrentStamina {
		return errors.New("not enough stamina")
	}

	unit.X = to.X
	unit.Y = to.Y
	unit.CurrentStamina -= totalCost
	gs.Players[accoID].Units[unitIndex] = unit

	// if player := gs.Players[accoID]; player != nil {
	// 	if i, ok := array.FindIndex(player.Units, func(p models.UnitInstance) bool {
	// 		return p.Unit.ID == unitID
	// 	}); ok {
	// 		// TODO: rules of stamina, range, obstacles

	// 		gs.Players[accoID].Units[i].X = to.X
	// 		gs.Players[accoID].Units[i].Y = to.Y
	// 	} else {
	// 		return errors.New("failed to find piece")
	// 	}
	// } else {
	// 	return errors.New("invalid Player")
	// }

	return nil
}

func (s *GameSession) EndTurn() {
	// if s.CurrentTurn == 1 {
	// 	s.CurrentTurn = 2
	// } else {
	// 	s.CurrentTurn = 1
	// }

	// // recarregar estamina
	// for i, piece := range s.Pieces {
	// 	if piece.Player == s.CurrentTurn {
	// 		s.Pieces[i].Piece.Stamina = 4 // ou outro valor baseado em item
	// 	}
	// }

	// s.BroadcastState()
}

func findPathToTarget(board []models.Cell, fromX, fromY, toX, toY int) []models.Position {
	// Implementar BFS com apenas 4 direções (N, S, E, W)
	// Garantir que não passa por fora do tabuleiro
	return nil // implementar depois
}

func findCell(board []models.Cell, x, y int) (models.Cell, bool) {
	for _, cell := range board {
		if cell.X == x && cell.Y == y {
			return cell, true
		}
	}
	return models.Cell{}, false
}

func generateBoard(sid string, elevationPercent float64) []models.Cell {
	if elevationPercent < 0 || elevationPercent > 1 {
		elevationPercent = 0.3
	}

	source := rand.NewSource(hashSid(sid))
	r := rand.New(source)

	cells := make([]models.Cell, 100)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			cells[y*10+x] = models.Cell{X: x, Y: y, Elevation: models.ElevationLow}
		}
	}
	totalElevated := int(float64(100) * elevationPercent)
	mediumCount := r.Intn(totalElevated + 1)
	highCount := totalElevated - mediumCount

	fillElevation(cells, r, models.ElevationMedium, mediumCount)
	fillElevation(cells, r, models.ElevationHigh, highCount)

	return cells
}

func fillElevation(cells []models.Cell, r *rand.Rand, elevation models.ElevationLevel, count int) {
	positions := r.Perm(100)
	placed := 0

	for _, pos := range positions {
		if cells[pos].Elevation == models.ElevationLow {
			cells[pos].Elevation = elevation
			placed++
			if placed == count {
				return
			}
		}
	}
}

func generateInitialPieces(sid string, unitis []models.Unit, used map[string]bool) []models.UnitInstance {
	source := rand.NewSource(hashSid(sid))
	r := rand.New(source)

	var piecePoints []models.UnitInstance

	for _, piece := range unitis {
		x, y := findAvailablePosition(r, used)
		key := positionToKey(x, y)
		used[key] = true

		piecePoints = append(piecePoints, models.UnitInstance{
			Position: models.Position{
				X: x,
				Y: y,
			},
			Unit:           piece,
			CurrentLife:    piece.LP,
			CurrentShield:  piece.Shield,
			CurrentStamina: piece.Stamina,
		})
	}

	return piecePoints
}

func buildUsedFromPiecePoints(piecePoints []models.UnitInstance) map[string]bool {
	used := make(map[string]bool)
	for _, pp := range piecePoints {
		key := positionToKey(pp.X, pp.Y)
		used[key] = true
	}
	return used
}

func findAvailablePosition(r *rand.Rand, used map[string]bool) (int, int) {

	for attempts := 0; attempts < 1000; attempts++ {
		x := r.Intn(10)
		y := r.Intn(10)
		if isPositionAvailable(x, y, used) {
			return x, y
		}
	}

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if isPositionAvailable(x, y, used) {
				return x, y
			}
		}
	}

	return 0, 0
}

func isPositionAvailable(x, y int, used map[string]bool) bool {
	key := positionToKey(x, y)
	if used[key] {
		return false
	}

	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < 10 && ny >= 0 && ny < 10 {
				if used[positionToKey(nx, ny)] {
					return false
				}
			}
		}
	}
	return true
}

func positionToKey(x, y int) string {
	return strconv.Itoa(x) + "-" + strconv.Itoa(y)
}

func hashSid(sid string) int64 {
	var hash int64
	for _, c := range sid {
		hash = hash*31 + int64(c)
	}
	return hash
}

func formatUUIDErrors(errorMap map[uuid.UUID]error) error {
	if len(errorMap) == 0 {
		return nil
	}

	var builder strings.Builder

	for id, err := range errorMap {
		if err == nil {
			continue
		}
		builder.WriteString(fmt.Sprintf("[%s]: %s\n", id.String(), err.Error()))
	}

	return errors.Errorf(strings.TrimSpace(builder.String()))
}
