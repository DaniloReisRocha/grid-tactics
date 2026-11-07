package game

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/DaniloReisRocha/grid-tactics/backend/models"
)

const minimumSpawnDistance = 5

// Chebyshev distance counts tile steps with the game's eight movement directions.
func spawnDistance(a, b models.Position) int {
	return max(abs(a.X-b.X), abs(a.Y-b.Y))
}

func safeSpawnCandidates(board []models.Cell, count int) []models.Position {
	region := largestRegion(board)
	// A couple of free neighbors in a tiny enclosed pocket are not enough.
	if len(region) < max(25, count*5) {
		return nil
	}
	cells := make(map[models.Position]models.Cell, len(region))
	for _, cell := range region {
		cells[models.Position{X: cell.X, Y: cell.Y}] = cell
	}
	candidates := []models.Position{}
	for _, cell := range region {
		// Start on solid ground, never on a bridge, in water or inside a structure.
		if cell.Terrain == "water" || cell.Object != "" {
			continue
		}
		position := models.Position{X: cell.X, Y: cell.Y}
		exits := 0
		for _, delta := range [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}} {
			next, exists := cells[models.Position{X: cell.X + delta[0], Y: cell.Y + delta[1]}]
			if exists && CanTraverse(cell, next) {
				exits++
			}
		}
		if exits >= 2 {
			candidates = append(candidates, position)
		}
	}
	return candidates
}

// Plan both teams together: a valid first team must not consume the second team's space.
func planSpawns(board []models.Cell, seed string, count int) ([]models.Position, error) {
	if count < 1 {
		return nil, errors.New("quantidade de spawns inválida")
	}
	candidates := safeSpawnCandidates(board, count)
	if len(candidates) < count {
		return nil, errors.New("área segura insuficiente para os personagens")
	}
	r := rand.New(rand.NewSource(int64(seedHash(seed))))
	order := r.Perm(len(candidates))
	// Farthest-point placement with deterministic restarts avoids unlucky greedy starts.
	for _, first := range order[:min(64, len(order))] {
		selected := []models.Position{candidates[first]}
		for len(selected) < count {
			bestIndex, bestDistance := -1, minimumSpawnDistance-1
			for _, index := range order {
				candidate := candidates[index]
				distance := int(^uint(0) >> 1)
				for _, placed := range selected {
					distance = min(distance, spawnDistance(candidate, placed))
				}
				if distance > bestDistance {
					bestIndex, bestDistance = index, distance
				}
			}
			if bestIndex < 0 {
				break
			}
			selected = append(selected, candidates[bestIndex])
		}
		if len(selected) == count {
			return selected, nil
		}
	}
	return nil, errors.New("não há posições seguras com distância mínima de 5 blocos")
}

func (s *GameSession) prepareSpawns(unitsPerPlayer int) error {
	if len(s.spawnSlots) > 0 {
		return nil
	}
	if len(s.Players) > 0 {
		return errors.New("posições iniciais não foram reservadas")
	}
	// Reserve the three-unit starter roster for every player before accepting the first.
	count := max(3, unitsPerPlayer) * s.maximumPlayers
	width, height := 0, 0
	for _, cell := range s.Board {
		width, height = max(width, cell.X+1), max(height, cell.Y+1)
	}
	if width < 1 || height < 1 || ((width+minimumSpawnDistance-1)/minimumSpawnDistance)*((height+minimumSpawnDistance-1)/minimumSpawnDistance) < count {
		return errors.New("mapa pequeno demais para separar os personagens por 5 blocos")
	}
	for attempt := 0; attempt < 64; attempt++ {
		seed := s.ID
		board := s.Board
		if attempt > 0 {
			seed = fmt.Sprintf("%s#spawn-%d", s.ID, attempt)
			board = generateBoard(seed, width, height)
		}
		positions, err := planSpawns(board, seed, count)
		if err == nil {
			s.Board, s.spawnSlots = board, positions
			return nil
		}
	}
	return errors.New("não foi possível criar um mapa com spawns seguros e separados por 5 blocos")
}
