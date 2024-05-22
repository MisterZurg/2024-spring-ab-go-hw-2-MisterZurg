package commands

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/config"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph"
)

func InitRunCommand() (cobra.Command, error) {
	cmdArgs := cmdargs.RunArgs{}
	cmd := cobra.Command{
		Use:   "run",
		Short: "Starts a leader election node",
		Long: `This command starts the leader election node that connects to zookeeper
		and starts to try to acquire leadership by creation of ephemeral node`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dg := depgraph.New()
			// zkConn, err := dg.GetZkConn()
			logger, err := dg.GetLogger()
			if err != nil {
				return fmt.Errorf("get logger: %w", err)
			}
			logger.Info("args received",
				slog.String("servers", strings.Join(cmdArgs.ZookeeperServers, ", ")),
				slog.Duration("leader-timeout", cmdArgs.LeaderTimeout),
				slog.Duration("attempter-timeout", cmdArgs.AttempterTimeout),
			)

			runner, err := dg.GetRunner()
			if err != nil {
				return fmt.Errorf("get runner: %w", err)
			}
			firstState, err := dg.GetEmptyState()
			if err != nil {
				return fmt.Errorf("get first state: %w", err)
			}
			err = runner.Run(cmd.Context(), firstState)
			if err != nil {
				return fmt.Errorf("run states: %w", err)
			}
			return nil
		},
	}

	// Get config from envs
	cfg, _ := config.New()
	// Configure parameters: flag -> env
	cmd.Flags().StringSliceVarP(&(cmdArgs.ZookeeperServers), "zk-servers", "s", []string{}, "Set the zookeeper servers.")
	if len(cmdArgs.ZookeeperServers) == 0 {
		cmdArgs.ZookeeperServers = cfg.ZKServers
	}

	cmd.Flags().DurationVarP(&(cmdArgs.LeaderTimeout), "leader-timeout", "l", 0, "Set the frequency at which the leader writes the file to disk.")
	if cmdArgs.LeaderTimeout == 0 {
		cmdArgs.LeaderTimeout = cfg.LeaderTimeout
	}

	cmd.Flags().DurationVarP(&(cmdArgs.AttempterTimeout), "attempter-timeout", "a", 0, "Set the frequency at which the attempter tries to beCUM a leader.")
	if cmdArgs.AttempterTimeout == 0 {
		cmdArgs.AttempterTimeout = cfg.AttempterTimeout
	}

	cmd.Flags().StringVarP(&(cmdArgs.FileDir), "file-dir", "f", "", "Set directory to write files on disk.")
	if cmdArgs.FileDir == "" {
		cmdArgs.FileDir = cfg.FileDir
	}

	cmd.Flags().IntVarP(&(cmdArgs.StorageCapacity), "storage-capacity", "c", 0, "Maximum files in 'file-dir'.")
	if cmdArgs.StorageCapacity == 0 {
		cmdArgs.StorageCapacity = cfg.StorageCapacity
	}

	return cmd, nil
}
