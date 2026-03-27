package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/rushteam/skills-cli/internal/config"
	"github.com/rushteam/skills-cli/internal/registry"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for available skill updates",
	RunE:  runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
	lock, err := config.LoadLock()
	if err != nil {
		return err
	}

	if len(lock.Skills) == 0 {
		fmt.Println(dimStyle.Render("No skills tracked in lock file."))
		fmt.Println(dimStyle.Render("Install skills with: skills-cli add <source>"))
		return nil
	}

	fmt.Println(textStyle.Render("Checking for skill updates..."))
	fmt.Println()

	cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	var updates []string
	var skipped []string
	var errors []string

	for name, entry := range lock.Skills {
		if entry.SkillFolderHash == "" || entry.SkillPath == "" || entry.Source == "" {
			skipped = append(skipped, name)
			continue
		}

		latestHash, err := registry.FetchSkillFolderHash(entry.Source, entry.SkillPath, "")
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
			continue
		}

		if latestHash != entry.SkillFolderHash {
			updates = append(updates, name)
		}
	}

	if len(updates) == 0 {
		fmt.Println(okStyle.Render("✓ All skills are up to date"))
	} else {
		fmt.Println(textStyle.Render(fmt.Sprintf("%d update(s) available:", len(updates))))
		fmt.Println()
		for _, name := range updates {
			fmt.Printf("  %s %s\n", cyanStyle.Render("↑"), name)
		}
		fmt.Println()
		fmt.Println(dimStyle.Render("Run ") + textStyle.Render("skills-cli update") + dimStyle.Render(" to update all skills"))
	}

	if len(skipped) > 0 {
		fmt.Println()
		fmt.Println(dimStyle.Render(fmt.Sprintf("%d skill(s) cannot be checked automatically:", len(skipped))))
		for _, name := range skipped {
			fmt.Printf("  %s %s\n", dimStyle.Render("•"), name)
		}
	}

	if len(errors) > 0 {
		fmt.Println()
		fmt.Println(dimStyle.Render(fmt.Sprintf("Could not check %d skill(s):", len(errors))))
		for _, e := range errors {
			fmt.Printf("  %s %s\n", dimStyle.Render("✗"), e)
		}
	}
	fmt.Println()
	return nil
}
