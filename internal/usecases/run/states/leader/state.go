package leader

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/go-zookeeper/zk"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states/init"
)

// countFiles — helper считающий кол-во файлов в dir
func countFiles(dir string) (int, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	return len(files), nil
}

// cleanDir — helper удаляющий файлы из dir
func cleanDir(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		errDelete := os.Remove(filepath.Join(dir, file.Name()))
		if errDelete != nil {
			return errDelete
		}
	}

	return nil
}

func New(runArgs cmdargs.RunArgs) *State {
	logger := runArgs.Logger.With("subsystem", "Leader")
	return &State{
		logger:        logger,
		leaderTimeout: *time.NewTicker(runArgs.LeaderTimeout),
	}
}

type State struct {
	logger  *slog.Logger
	runArgs cmdargs.RunArgs

	connection *zk.Conn
	// Для Gracefull Shutup
	leaderTimeout time.Ticker
	states        init.Stater
}

func (s *State) WithConnection(connection *zk.Conn) *State {
	s.connection = connection
	return s
}

func (s *State) GracefullShutup() {
	s.leaderTimeout.Stop()
}

func (s *State) String() string {
	return "LeaderState"
}

// Run для LeaderState — cтали лидером, нужно писать файлик на диск(симуляция полезной деятельности)
// — Если отвалилась жепа и стал недоступен Zookeper, переходим в состояние Failover
// — Если SIGTERM aka ctx.Done в состояние Stopping
func (s *State) Run(ctx context.Context) (states.AutomataState, error) {
	if s.connection == nil {
		return s.states.GetFailoverState(s.runArgs)
	}
	// Стали лидером, нужно писать файлик на диск(симуляция полезной деятельности)
	// Реплика, которая становится лидером,
	// должна каждые leader-timeout секунд писать файл в директорию file-dir,
	result := make(chan error)
	// leaderTimeout := time.NewTicker(s.runArgs.LeaderTimeout)

	go func() {
		for range s.leaderTimeout.C {
			if s.connection.State() != zk.StateHasSession {
				result <- nil
				return
			}

			fileNumber, err := countFiles(s.runArgs.FileDir)
			if err != nil {
				result <- err
				return
			}
			// а также удалять старые файлы, если количество файлов в директории больше,
			// чем storage-capacity
			if s.runArgs.StorageCapacity < fileNumber {
				errClean := cleanDir(s.runArgs.FileDir)
				if errClean != nil {
					result <- errClean
					return
				}
			}

			fileName := fmt.Sprintf("%s_%s.txt", time.RFC850, time.Now())
			filePath := filepath.Join(s.runArgs.FileDir, fileName)
			_, errCreateFile := os.Create(filePath)
			if errCreateFile != nil {
				result <- errCreateFile
				break
			}
		}
	}()

	select {
	case <-ctx.Done():
		return s.states.GetStoppingState(s.runArgs)
	case err := <-result:
		if err != nil {
			s.logger.LogAttrs(ctx, slog.LevelError, fmt.Sprintf("Error from leader file system in directory %s: %v", s.runArgs.FileDir, err))
			return s.states.GetStoppingState(s.runArgs)
		}

		return s.states.GetFailoverState(s.runArgs)
	}
}
