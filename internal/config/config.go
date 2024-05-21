package config

import (
	"github.com/caarlos0/env/v11"
	"time"
)

type Config struct {
	ZKServers        []string      `env:"ELECTION_ZK_SERVERS" envSeparator:"," envDefault:"foo1.bar:2181,foo2.bar:2181"`
	LeaderTimeout    time.Duration `env:"ELECTION_LEADER_TIMEOUT" envDefault:"10s"`
	AttempterTimeout time.Duration `env:"ELECTION_ATTEMPTER_TIMEOUT" envDefault:"10s"`
	FileDir          string        `env:"ELECTION_FILE_DIR" envDefault:"/tmp/election"`
	StorageCapacity  int           `env:"ELECTION_STOEAGE_CAPACITY" envDefault:"10"`
}

func New() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
