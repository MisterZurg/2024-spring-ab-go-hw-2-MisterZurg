package failover

import (
	"context"
	"errors"
	"fmt"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/init"
	"github.com/go-zookeeper/zk"
	"log/slog"
	"time"
)

func New(runArgs cmdargs.RunArgs) *State {
	logger := runArgs.Logger.With("subsystem", "Failover")
	return &State{
		logger: logger,
	}
}

type State struct {
	logger *slog.Logger

	runArgs cmdargs.RunArgs

	states init.Stater
}

func (s *State) String() string {
	return "FailoverState"
}

type Result struct {
	connection *zk.Conn
	err        error
}

func (s *State) unfuckFailoverState(ctx context.Context, result chan Result) {
	const maxAttempts = 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		connection, _, err := zk.Connect(
			s.runArgs.ZookeeperServers,
			s.runArgs.SessionTimeout,
		)
		if err == nil {
			result <- Result{
				connection: connection,
			}
			return
		}
		s.logger.LogAttrs(ctx, slog.LevelError, fmt.Sprintf("error while re-connecting to Zookeeper on attempt %d", attempt))

		time.Sleep(s.runArgs.AttempterTimeout)
	}
	result <- Result{
		err: errors.New("unable to re-connect to Zookeeper"),
	}
}

// Run для FailoverState — попытка приложения починить самого себя
func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	result := make(chan Result)
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
