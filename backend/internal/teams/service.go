package teams

import (
	"context"
	"fmt"

	"github.com/racecontrol/backend/internal/db/dbgen"
)

type Team struct {
	ID           int64
	Code         string
	Name         string
	CategoryCode string
	CategoryName string
}

type Store interface {
	ListTeams(ctx context.Context) ([]dbgen.ListTeamsRow, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context) ([]Team, error) {
	rows, err := s.store.ListTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}
	teams := make([]Team, 0, len(rows))
	for _, row := range rows {
		teams = append(teams, Team(row))
	}
	return teams, nil
}
