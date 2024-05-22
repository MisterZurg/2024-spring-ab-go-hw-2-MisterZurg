package init

import (
	"context"
	"log/slog"

	"github.com/go-zookeeper/zk"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/attempter"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/failover"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/leader"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/stopping"
)

type Stater interface {
	GetAttempterState(cmdargs.RunArgs) (*attempter.State, error)
	GetLeaderState(cmdargs.RunArgs) (*leader.State, error)
	GetFailoverState(cmdargs.RunArgs) (*failover.State, error)
	GetStoppingState(cmdargs.RunArgs) (*stopping.State, error)
}

func New(runArgs cmdargs.RunArgs) *State {
	logger := runArgs.Logger.With("subsystem", "Init")
	return &State{
		logger: logger,
	}
}

type State struct {
	logger  *slog.Logger
	runArgs cmdargs.RunArgs

	states Stater
}

func (s *State) String() string {
	return "InitState"
}

type Result struct {
	conn *zk.Conn
	err  error
}

// Run для InitState — начинается инициализация, проверка доступности всех ресурсов
// — Если инициализация успешна, переходим в состояние Attempter
// — Если отвалилась жепа и стал недоступен Zookeper, переходим в состояние Failover
// — Если SIGTERM aka ctx.Done в состояние Stopping
func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Inniting State")

	result := make(chan Result)
	go func() {
		connection, _, err := zk.Connect(
			s.runArgs.ZookeeperServers,
			s.runArgs.SessionTimeout,
		)

		result <- Result{
			conn: connection,
			err:  err,
		}
	}()

	// Логика из грфа состо
	select {
	case <-ctx.Done():
		return s.states.GetFailoverState(s.runArgs)
	case res := <-result:
		if res.err != nil {
			s.logger.LogAttrs(ctx, slog.LevelError, "can not connect to zookeeper", slog.String("msg", res.err.Error()))

			return s.states.GetFailoverState(s.runArgs)
		}
		attempter, err := s.states.GetAttempterState(s.runArgs)
		if err != nil {
			return nil, err
		}

		return attempter.WithConnection(res.conn), nil
	}
}
