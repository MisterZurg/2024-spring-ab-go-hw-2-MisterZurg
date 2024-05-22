package states

import (
	"context"
	"errors"
	"fmt"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/go-zookeeper/zk"
	"log/slog"
	"time"
)

func NewAttempterState(runArgs cmdargs.RunArgs) *AttempterState {
	logger := runArgs.Logger.With("subsystem", "Attempter")
	return &AttempterState{
		logger:           logger,
		attempterTimeout: *time.NewTicker(runArgs.AttempterTimeout),
	}
}

type AttempterState struct {
	logger  *slog.Logger
	runArgs cmdargs.RunArgs

	connection *zk.Conn
	// Для Gracefull Shutup
	attempterTimeout time.Ticker
	states           Stater
}

func (s *AttempterState) WithConnection(connection *zk.Conn) *AttempterState {
	s.connection = connection
	return s
}

func (s *AttempterState) GracefullShutup() {
	s.attempterTimeout.Stop()
}

func (s *AttempterState) String() string {
	return "AttempterState"
}

// Run для AttempterState — пытаемся стать лидером - раз в attempter-timeout пытаемся создать эфемерную ноду в зукипере
// — Если отвалилась жепа и стал недоступен Zookeper, переходим в состояние Failover
// — Если удаецца создать эфемерную ноду в Zookeper, переходим в Leader
// — Если SIGTERM aka ctx.Done в состояние Stopping
func (s *AttempterState) Run(ctx context.Context) (AutomataState, error) {
	if s.connection == nil {
		return s.states.GetFailoverState(s.runArgs)
	}

	result := make(chan error)
	// Пытаемся стать лидером - раз в attempter-timeout пытаемся создать эфемерную ноду в зукипере
	go func() {
		for range s.attempterTimeout.C {
			_, err := s.connection.Create(
				s.runArgs.FileDir,
				[]byte(fmt.Sprintf("time: %s", time.Now())),
				zk.FlagEphemeral,
				zk.WorldACL(zk.PermRead),
			)
			// This function is used to check if the error err is equal to
			// the target error or if it satisfies the error interface
			if !errors.As(err, zk.ErrNodeExists) {
				result <- err
				return
			}

			s.logger.LogAttrs(ctx, slog.LevelDebug, "failed attempt: node is already exist")
		}
	}()

	select {
	case <-ctx.Done():
		return s.states.GetStoppingState(s.runArgs)

	case err := <-result:
		if err != nil {
			s.logger.LogAttrs(ctx, slog.LevelError, "shit occurred", slog.String("msg", err.Error()))
			return s.states.GetFailoverState(s.runArgs)
		}

		leader, err := s.states.GetLeaderState(s.runArgs)
		if err != nil {
			return nil, err
		}

		return leader.WithConnection(s.connection), nil
	}
}
