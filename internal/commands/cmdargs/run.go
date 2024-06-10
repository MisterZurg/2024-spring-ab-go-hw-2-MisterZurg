package cmdargs

import (
	"log/slog"
	"time"
)

type RunArgs struct {
	ZookeeperServers []string
	LeaderTimeout    time.Duration
	AttempterTimeout time.Duration
	SessionTimeout   time.Duration
	FileDir          string
	StorageCapacity  int
	Logger           *slog.Logger
}
