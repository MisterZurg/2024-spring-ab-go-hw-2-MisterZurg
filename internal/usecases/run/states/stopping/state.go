package stopping

import (
	"context"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/init"
	"log/slog"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

func New(runArgs cmdargs.RunArgs) *State {
	logger := runArgs.Logger.With("subsystem", "Stopping")
	return &State{
		logger: logger,
	}
}

type State struct {
	logger  *slog.Logger
	runArgs cmdargs.RunArgs

	states init.Stater
}

func (s *State) String() string {
	return "StoppingState"
}

// Run для StoppingState — Graceful shutdown
func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
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

	return nil, nil
}
