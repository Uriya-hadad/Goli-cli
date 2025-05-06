package teamFunctions

import (
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	"goli-cli/cli/teamFunctions/ops"
	"goli-cli/cli/teamFunctions/performance"
	"goli-cli/cli/teamFunctions/shanyTeam"
	"goli-cli/cli/teamFunctions/shared"
	"goli-cli/entities"
	"goli-cli/types"
	"sync"
)

const (
	Performance = "performance"
	Ops         = "ops"
	ShanyTeam   = "shanyTeam"
)

func NewCmd(cf *client.Client, role string, apps **map[string]types.AppData, instances **map[string][]*entities.Instance, offerNames **[]string, appsLock, instancesLock, updateDataLock *sync.WaitGroup) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "team-features",
		Aliases: []string{"tf", "features"},
		Short:   "Access role-specific features for advanced application and instances operations.",
		Long: `The 'team-features' command is a role-specific tool that provides access to advanced features tailored to your team and responsibilities.
Running this command opens an interactive mode where you can view and select from the available sub-commands, which are dynamically assigned based on your role.

Since the sub-commands are role-dependent, they may vary between users.
This ensures you see only the tools relevant to your specific operational needs.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			instancesLock.Wait()
			appsLock.Wait()
		},
		// TODO clashes with line 63.. need to rethink about that
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			TeamFeaturesCli(cf, role, *apps, *instances, *offerNames, updateDataLock)
		},
	}
	generateSubCommands(cmd, role, cf, apps, instances)
	return cmd
}

func TeamFeaturesCli(cf *client.Client, role string, apps *map[string]types.AppData, instances *map[string][]*entities.Instance, offerNames *[]string, updateDataLock *sync.WaitGroup) {
	switch role {
	case Performance:
		performance.PerformanceCli(cf, apps, instances, updateDataLock)
	case Ops:
		ops.OpsCli(cf, apps, instances, offerNames)
	case ShanyTeam:
		shanyTeam.ShanyTeamCli(cf, apps, instances, offerNames)
	default:
		shared.SharedCli(cf, apps, instances, offerNames)
	}

}

func generateSubCommands(cmd *cobra.Command, role string, cf *client.Client, apps **map[string]types.AppData, instances **map[string][]*entities.Instance) {
	switch role {
	case Performance:
		cmd.AddCommand(performance.NewStatusCmd(cf, apps))
	case Ops:
		cmd.AddCommand(ops.NewRunSaaCmd(cf, apps, instances),
			ops.NewSaaStatusCmd(cf, instances),
			ops.NewRunQueryAllCmd(instances),
			ops.NewCheckCertCmd(cf, apps, instances))
	}

	// shared commands
	cmd.AddCommand(shared.NewRunQueryCmd(cf, instances),
		shared.NewConnectToDbCmd(cf, instances))

}
