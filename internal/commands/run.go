package commands

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph"
	"github.com/spf13/cobra"
)

func InitRunCommand() (cobra.Command, error) {
	cmdArgs := cmdargs.RunArgs{}
	cmd := cobra.Command{
		Use:   "run",
		Short: "Starts a leader election node",
		Long: `This command starts the leader election node that connects to zookeeper
		and starts to try to acquire leadership by creation of ephemeral node`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dg := depgraph.New()
			// zkConn, err := dg.GetZkConn()
			logger, err := dg.GetLogger()
			if err != nil {
				return fmt.Errorf("get logger: %w", err)
			}
			logger.Info("args received", slog.String("servers", strings.Join(cmdArgs.ZookeeperServers, ", ")))

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

	cmd.Flags().StringSliceVarP(&(cmdArgs.ZookeeperServers), "zk-servers", "s", []string{}, "Set the zookeeper servers.")

	return cmd, nil
}
