package cmd

import (
	"fmt"

	"github.com/rushteam/skills-cli/internal/config"
	syncer "github.com/rushteam/skills-cli/internal/sync"
	"github.com/spf13/cobra"
)

var (
	pushAgent   []string
	pushProject []string
	pushAll     bool
	pushForce   bool
	pushDiff    bool
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push skills from central store to agent directories",
	Long: `Push skills from the central skills store (~/.skills-cli/skills/) to
global agent directories and/or project-level agent directories.`,
	RunE: runPush,
}

func init() {
	pushCmd.Flags().StringSliceVarP(&pushAgent, "agent", "a", nil, "Agent(s) to push to")
	pushCmd.Flags().StringSliceVarP(&pushProject, "project", "p", nil, "Project path(s) to push to")
	pushCmd.Flags().BoolVar(&pushAll, "all", false, "Push to all registered projects and global agents")
	pushCmd.Flags().BoolVar(&pushForce, "force", false, "Skip conflict prompts, overwrite directly")
	pushCmd.Flags().BoolVar(&pushDiff, "diff", false, "Show diff only, do not sync (dry-run)")
}

func runPush(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	targets := syncer.ResolveTargets(cfg, pushAgent, pushProject, pushAll)
	if len(targets) == 0 {
		fmt.Println(dimStyle.Render("No targets to push to."))
		fmt.Println(dimStyle.Render("Use --agent, --project, or --all to specify destinations."))
		fmt.Println(dimStyle.Render("Or configure default_agents in: ") + textStyle.Render(config.ConfigPath()))
		return nil
	}

	showLogo()
	fmt.Println()
	fmt.Println(titleStyle.Render("  Pushing skills from central store"))
	fmt.Println()

	return syncer.Push(targets, syncer.SyncOptions{
		Force:    pushForce,
		DiffOnly: pushDiff,
	})
}
