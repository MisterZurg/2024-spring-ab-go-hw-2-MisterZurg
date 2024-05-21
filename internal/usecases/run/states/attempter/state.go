package attempter

import (
	"context"
	"log/slog"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

func New(logger *slog.Logger) *State {
	logger = logger.With("subsystem", "Attempter")
	return &State{
		logger: logger,
	}
}

type State struct {
	logger *slog.Logger
}

func (s *State) String() string {
	return "AttempterState"
}

func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	// TODO add logic:
	// Пытаемся стать лидером - раз в attempter-timeout пытаемся создать эфемерную ноду в зукипере
	return nil, nil
}
