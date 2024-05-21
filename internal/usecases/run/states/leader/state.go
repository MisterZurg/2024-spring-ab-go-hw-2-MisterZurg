package leader

import (
	"context"
	"log/slog"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

func New(logger *slog.Logger) *State {
	logger = logger.With("subsystem", "Leader")
	return &State{
		logger: logger,
	}
}

type State struct {
	logger *slog.Logger
}

func (s *State) String() string {
	return "LeaderState"
}

func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	// TODO add logic:
	// Стали лидером, нужно писать файлик на диск(симуляция полезной деятельности)
	// Реплика, которая становится лидером,
	// должна каждые leader-timeout секунд писать файл в директорию file-dir,
	// а также удалять старые файлы, если количество файлов в директории больше,
	// чем storage-capacity
	return nil, nil
}
