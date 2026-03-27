package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/rushteam/skills-cli/internal/agent"
	"github.com/rushteam/skills-cli/internal/config"
	sk "github.com/rushteam/skills-cli/internal/skill"
	"github.com/spf13/cobra"
)

var (
	removeGlobal bool
	removeAgent  []string
	removeYes    bool
	removeAll    bool
)

var removeCmd = &cobra.Command{
	Use:     "remove [skills...]",
	Aliases: []string{"rm"},
	Short:   "Remove installed skills",
	RunE:    runRemove,
}

func init() {
	removeCmd.Flags().BoolVarP(&removeGlobal, "global", "g", false, "Remove from global scope")
	removeCmd.Flags().StringSliceVarP(&removeAgent, "agent", "a", nil, "Remove from specific agent(s)")
	removeCmd.Flags().BoolVarP(&removeYes, "yes", "y", false, "Skip confirmation")
	removeCmd.Flags().BoolVar(&removeAll, "all", false, "Remove all skills")
}

func runRemove(cmd *cobra.Command, args []string) error {
	centralDir := config.SkillsHome()
	okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	skillNames, err := agent.ScanSkillsInDir(centralDir)
	if err != nil || len(skillNames) == 0 {
		fmt.Println(dimStyle.Render("No skills installed."))
		return nil
	}

	var toRemove []string

	if removeAll {
		toRemove = skillNames
	} else if len(args) > 0 {
		toRemove = args
	} else {
		var options []huh.Option[string]
		for _, n := range skillNames {
			options = append(options, huh.NewOption(n, n))
		}

		var selected []string
		err := huh.NewMultiSelect[string]().
			Title("Select skills to remove:").
			Options(options...).
			Value(&selected).
			Run()
		if err != nil || len(selected) == 0 {
			fmt.Println(dimStyle.Render("Cancelled"))
			return nil
		}
		toRemove = selected
	}

	if !removeYes && !removeAll {
		var confirm bool
		err := huh.NewConfirm().
			Title(fmt.Sprintf("Remove %d skill(s)?", len(toRemove))).
			Value(&confirm).
			Run()
		if err != nil || !confirm {
			fmt.Println(dimStyle.Render("Cancelled"))
			return nil
		}
	}

	lock, _ := config.LoadLock()

	for _, name := range toRemove {
		dir := filepath.Join(centralDir, name)
		if err := sk.RemoveSkillDir(dir); err != nil && !os.IsNotExist(err) {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(
				fmt.Sprintf("  ✗ %s: %v", name, err)))
			continue
		}
		lock.RemoveSkill(name)
		fmt.Println(okStyle.Render(fmt.Sprintf("  ✓ removed %s", name)))
	}

	config.SaveLock(lock)
	fmt.Println()
	return nil
}
