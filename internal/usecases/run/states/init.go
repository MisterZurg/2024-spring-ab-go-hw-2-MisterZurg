package states

import (
	"context"
	"log/slog"

	"github.com/go-zookeeper/zk"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
)

type Stater interface {
	GetAttempterState(cmdargs.RunArgs) (*AttempterState, error)
	GetLeaderState(cmdargs.RunArgs) (*LeaderState, error)
	GetFailoverState(cmdargs.RunArgs) (*FailoverState, error)
	GetStoppingState(cmdargs.RunArgs) (*StoppingState, error)
}

func NewInitState(runArgs cmdargs.RunArgs) *InitState {
	logger := runArgs.Logger.With("subsystem", "Init")
	return &InitState{
		logger: logger,
	}
}

type InitState struct {
	logger  *slog.Logger
	runArgs cmdargs.RunArgs

	states Stater
}

func (s *InitState) String() string {
	return "InitState"
}

type ResultInitState struct {
	conn *zk.Conn
	err  error
}

// Run для InitState — начинается инициализация, проверка доступности всех ресурсов
// — Если инициализация успешна, переходим в состояние Attempter
// — Если отвалилась жепа и стал недоступен Zookeper, переходим в состояние Failover
// — Если SIGTERM aka ctx.Done в состояние Stopping
func (s *InitState) Run(ctx context.Context) (AutomataState, error) {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Inniting State")

	result := make(chan ResultInitState)
	go func() {
		connection, _, err := zk.Connect(
			s.runArgs.ZookeeperServers,
			s.runArgs.SessionTimeout,
		)

		result <- ResultInitState{
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
