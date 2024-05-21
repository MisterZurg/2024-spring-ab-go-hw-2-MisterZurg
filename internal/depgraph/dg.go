package depgraph

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/attempter"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/empty"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/failover"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/init"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/leader"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/stopping"
)

type dgEntity[T any] struct {
	sync.Once
	value   T
	initErr error
}

func (e *dgEntity[T]) get(init func() (T, error)) (T, error) {
	e.Do(func() {
		e.value, e.initErr = init()
	})
	if e.initErr != nil {
		return *new(T), e.initErr
	}
	return e.value, nil
}

type DepGraph struct {
	logger      *dgEntity[*slog.Logger]
	stateRunner *dgEntity[*run.LoopRunner]

	initState      *dgEntity[*init.State]
	leaderState    *dgEntity[*leader.State]
	attempterState *dgEntity[*attempter.State]
	failoverState  *dgEntity[*failover.State]
	stoppingState  *dgEntity[*stopping.State]

	emptyState *dgEntity[*empty.State]
}

func New() *DepGraph {
	return &DepGraph{
		logger:      &dgEntity[*slog.Logger]{},
		stateRunner: &dgEntity[*run.LoopRunner]{},

		initState:      &dgEntity[*init.State]{},
		leaderState:    &dgEntity[*leader.State]{},
		attempterState: &dgEntity[*attempter.State]{},
		failoverState:  &dgEntity[*failover.State]{},
		stoppingState:  &dgEntity[*stopping.State]{},

		emptyState: &dgEntity[*empty.State]{},
	}
}

func (dg *DepGraph) GetLogger() (*slog.Logger, error) {
	return dg.logger.get(func() (*slog.Logger, error) {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})), nil
	})
}

// GetEmptyState - example стейт для примера, в итоговом сервисе использоваться не должен
func (dg *DepGraph) GetEmptyState() (*empty.State, error) {
	return dg.emptyState.get(func() (*empty.State, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("get logger: %w", err)
		}
		return empty.New(logger), nil
	})
}

func (dg *DepGraph) GetLeaderState() (*leader.State, error) {
	return dg.leaderState.get(func() (*leader.State, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("get logger: %w", err)
		}
		return leader.New(logger), nil
	})
}

func (dg *DepGraph) GetRunner() (run.Runner, error) {
	return dg.stateRunner.get(func() (*run.LoopRunner, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("get logger: %w", err)
		}
		return run.NewLoopRunner(logger), nil
	})
}
