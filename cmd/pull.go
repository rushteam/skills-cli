package cmd

import (
	"fmt"

	"github.com/rushteam/skills-cli/internal/config"
	syncer "github.com/rushteam/skills-cli/internal/sync"
	"github.com/spf13/cobra"
)

var (
	pullAgent   []string
	pullProject []string
	pullAll     bool
	pullForce   bool
	pullDiff    bool
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull skills from agent directories to central store",
	Long: `Pull skills from global agent directories and/or project-level agent directories
into the central skills store (~/.skills-cli/skills/).`,
	RunE: runPull,
}

func init() {
	pullCmd.Flags().StringSliceVarP(&pullAgent, "agent", "a", nil, "Agent(s) to pull from")
	pullCmd.Flags().StringSliceVarP(&pullProject, "project", "p", nil, "Project path(s) to pull from")
	pullCmd.Flags().BoolVar(&pullAll, "all", false, "Pull from all registered projects and global agents")
	pullCmd.Flags().BoolVar(&pullForce, "force", false, "Skip conflict prompts, overwrite directly")
	pullCmd.Flags().BoolVar(&pullDiff, "diff", false, "Show diff only, do not sync (dry-run)")
}

func runPull(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	targets := syncer.ResolveTargets(cfg, pullAgent, pullProject, pullAll)
	if len(targets) == 0 {
		fmt.Println(dimStyle.Render("No targets to pull from."))
		fmt.Println(dimStyle.Render("Use --agent, --project, or --all to specify sources."))
		fmt.Println(dimStyle.Render("Or configure default_agents in: ") + textStyle.Render(config.ConfigPath()))
		return nil
	}

	showLogo()
	fmt.Println()
	fmt.Println(titleStyle.Render("  Pulling skills to central store"))
	fmt.Println()

	return syncer.Pull(targets, syncer.SyncOptions{
		Force:    pullForce,
		DiffOnly: pullDiff,
	})
}
