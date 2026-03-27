package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/rushteam/skills-cli/internal/agent"
	"github.com/rushteam/skills-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	listAgent   string
	listGlobal  bool
	listProject string
	listJSON    bool
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List installed skills",
	RunE:    runList,
}

func init() {
	listCmd.Flags().StringVarP(&listAgent, "agent", "a", "", "Filter by agent name")
	listCmd.Flags().BoolVarP(&listGlobal, "global", "g", false, "List only global agent skills")
	listCmd.Flags().StringVarP(&listProject, "project", "p", "", "List only skills from a specific project")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var skills []agent.SkillInfo

	if listGlobal {
		skills, err = agent.ListGlobalSkills(cfg)
	} else if listProject != "" {
		projCfg := &config.Config{
			Agents: cfg.Agents,
			Projects: []config.ProjectConfig{
				{Path: listProject},
			},
		}
		skills, err = agent.ListProjectSkills(projCfg)
	} else {
		skills, err = agent.ListAllSkills(cfg)
	}
	if err != nil {
		return err
	}

	if listAgent != "" {
		var filtered []agent.SkillInfo
		for _, s := range skills {
			if s.Agent == listAgent || s.Scope == "central" {
				filtered = append(filtered, s)
			}
		}
		skills = filtered
	}

	if listJSON {
		data, _ := json.MarshalIndent(skills, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if len(skills) == 0 {
		fmt.Println(dimStyle.Render("No skills found."))
		fmt.Println(dimStyle.Render("Install skills with: skills-cli add <source>"))
		return nil
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	scopeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Bold(true)
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	fmt.Println()
	fmt.Println(headerStyle.Render(fmt.Sprintf("  Found %d skill(s)", len(skills))))
	fmt.Println()

	currentScope := ""
	currentProject := ""
	for _, s := range skills {
		scopeLabel := s.Scope
		if s.Scope == "project" {
			scopeLabel = fmt.Sprintf("project:%s", agent.ShortenPath(s.Project))
		}

		if scopeLabel != currentScope || s.Project != currentProject {
			if currentScope != "" {
				fmt.Println()
			}
			fmt.Println(scopeStyle.Render(fmt.Sprintf("  [%s]", scopeLabel)))
			currentScope = scopeLabel
			currentProject = s.Project
		}

		agentLabel := ""
		if s.Agent != "" {
			agentLabel = dimStyle.Render(fmt.Sprintf(" (%s)", s.Agent))
		}

		shortPath := agent.ShortenPath(s.Path)
		fmt.Printf("    %s%s\n", nameStyle.Render(s.Name), agentLabel)
		fmt.Printf("    %s\n", pathStyle.Render(shortPath))
	}
	fmt.Println()

	return nil
}
