package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var initSkillCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new SKILL.md template",
	Long:  `Initialize a new skill directory with a SKILL.md template.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInitSkill,
}

func runInitSkill(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	skillName := filepath.Base(cwd)
	hasName := false

	if len(args) > 0 {
		skillName = args[0]
		hasName = true
	}

	skillDir := cwd
	if hasName {
		skillDir = filepath.Join(cwd, skillName)
	}
	skillFile := filepath.Join(skillDir, "SKILL.md")
	displayPath := "SKILL.md"
	if hasName {
		displayPath = filepath.Join(skillName, "SKILL.md")
	}

	okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	if _, err := os.Stat(skillFile); err == nil {
		fmt.Println(textStyle.Render(fmt.Sprintf("Skill already exists at %s", dimStyle.Render(displayPath))))
		return nil
	}

	if hasName {
		os.MkdirAll(skillDir, 0o755)
	}

	content := fmt.Sprintf(`---
name: %s
description: A brief description of what this skill does
---

# %s

Instructions for the agent to follow when this skill is activated.

## When to use

Describe the scenarios where this skill should be used.

## Instructions

1. First step
2. Second step
3. Additional steps as needed
`, skillName, skillName)

	if err := os.WriteFile(skillFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write SKILL.md: %w", err)
	}

	fmt.Println(okStyle.Render(fmt.Sprintf("Initialized skill: %s", skillName)))
	fmt.Println()
	fmt.Println(dimStyle.Render("Created:"))
	fmt.Printf("  %s\n", displayPath)
	fmt.Println()
	fmt.Println(dimStyle.Render("Next steps:"))
	fmt.Printf("  1. Edit %s to define your skill instructions\n", textStyle.Render(displayPath))
	fmt.Printf("  2. Update the %s and %s in the frontmatter\n", textStyle.Render("name"), textStyle.Render("description"))
	fmt.Println()
	fmt.Println(dimStyle.Render("Publishing:"))
	fmt.Printf("  Push to a repo, then %s\n", textStyle.Render("skills-cli add <owner>/<repo>"))
	fmt.Println()
	return nil
}
