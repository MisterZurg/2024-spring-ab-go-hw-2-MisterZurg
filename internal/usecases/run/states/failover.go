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

func NewFailoverState(runArgs cmdargs.RunArgs) *FailoverState {
	logger := runArgs.Logger.With("subsystem", "Failover")
	return &FailoverState{
		logger: logger,
	}
}

type FailoverState struct {
	logger *slog.Logger

	runArgs cmdargs.RunArgs

	states Stater
}

func (s *FailoverState) String() string {
	return "FailoverState"
}

type ResultFailover struct {
	connection *zk.Conn
	err        error
}

func (s *FailoverState) unfuckFailoverState(ctx context.Context, result chan ResultFailover) {
	const maxAttempts = 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		connection, _, err := zk.Connect(
			s.runArgs.ZookeeperServers,
			s.runArgs.SessionTimeout,
		)
		if err == nil {
			result <- ResultFailover{
				connection: connection,
			}
			return
		}
		s.logger.LogAttrs(ctx, slog.LevelError, fmt.Sprintf("error while re-connecting to Zookeeper on attempt %d", attempt))

		time.Sleep(s.runArgs.AttempterTimeout)
	}
	result <- ResultFailover{
		err: errors.New("unable to re-connect to Zookeeper"),
	}
}

// Run для FailoverState — попытка приложения починить самого себя
func (s *FailoverState) Run(ctx context.Context) (AutomataState, error) {
	result := make(chan ResultFailover)
	go s.unfuckFailoverState(ctx, result)

	select {
	case <-ctx.Done():
		return s.states.GetStoppingState(s.runArgs)
	case res := <-result:
		if res.err != nil {
			return s.states.GetStoppingState(s.runArgs)
		}

		attempter, err := s.states.GetAttempterState(s.runArgs)
		if err != nil {
			return nil, err
		}
		return attempter.WithConnection(res.connection), nil
	}
}
