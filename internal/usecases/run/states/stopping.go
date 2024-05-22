package states

import (
	"context"
	"log/slog"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
)

func NewStoppingState(runArgs cmdargs.RunArgs) *StoppingState {
	logger := runArgs.Logger.With("subsystem", "Stopping")
	return &StoppingState{
		logger: logger,
	}
}

type StoppingState struct {
	logger  *slog.Logger
	runArgs cmdargs.RunArgs

	states Stater
}

func (s *StoppingState) String() string {
	return "StoppingState"
}

// Run для StoppingState — Graceful shutdown
func (s *StoppingState) Run(ctx context.Context) (AutomataState, error) {
	// Graceful shutdown - состояние, в котором приложение оСВОбождает все СВОи ресурсы
	attempter, err := s.states.GetAttempterState(s.runArgs)
	if err != nil {
		return nil, err
	}

	leader, err := s.states.GetLeaderState(s.runArgs)
	if err != nil {
		return nil, err
	}

	attempter.GracefullShutup()
	leader.GracefullShutup()

	s.logger.LogAttrs(ctx, slog.LevelWarn, "server is stopped Gracefully")

	return nil, nil //nolint:nilnil // stopped Gracefully nothing to return
}
