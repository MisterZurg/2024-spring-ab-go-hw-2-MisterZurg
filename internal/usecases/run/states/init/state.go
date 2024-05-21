package init

import (
	"context"
	"log/slog"

	"github.com/go-zookeeper/zk"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

type Stater interface {
	GetAttempterState()
	GetLeaderState()
	GetFailoverState()
	GetStoppingState()
}

func New(logger *slog.Logger) *State {
	logger = logger.With("subsystem", "Init")
	return &State{
		logger: logger,
	}
}

type State struct {
	logger *slog.Logger
	states Stater
}

func (s *State) String() string {
	return "InitState"
}

type Result struct {
	conn *zk.Conn
	err  error
}

// Run для InitState
// - Начинается инициализация, проверка доступности всех ресурсов
func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Inniting State")

	result := make(chan Result)
	go func() {
		connection, _, err := zk.Connect()

		result <- Result{
			conn: connection,
			err:  err,
		}
	}()

	select {
	case <-ctx.Done():
		return s.states.GetFailoverState()
	}
}
