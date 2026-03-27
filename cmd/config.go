package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rushteam/skills-cli/internal/agent"
	"github.com/rushteam/skills-cli/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage skills-cli configuration",
}

// config show
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Println(titleStyle.Render("  Configuration") + " " + dimStyle.Render(config.ConfigPath()))
		fmt.Println()
		fmt.Println(string(data))
		return nil
	},
}

// config init
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
		cfg := config.DefaultConfig()
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println(okStyle.Render("✓ Configuration initialized"))
		fmt.Println(dimStyle.Render("  " + config.ConfigPath()))
		return nil
	},
}

// config add-agent
var (
	addAgentName        string
	addAgentProjectPath string
	addAgentGlobalPath  string
)

var configAddAgentCmd = &cobra.Command{
	Use:   "add-agent [name] [project-path]",
	Short: "Add a custom agent to ~/.skills-cli/config.yaml",
	Long: `Add or overwrite an agent entry under the agents map in the config file.

Provide name and project-level skills path (relative to repo root). If --global-path
is omitted, it defaults to ~/<project-path> (same pattern as built-in agents).

Examples:
  skills-cli config add-agent my-tool .my-tool/skills
  skills-cli config add-agent my-tool .my-tool/skills --global-path ~/.config/my-tool/skills`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := addAgentName
		proj := addAgentProjectPath
		switch len(args) {
		case 2:
			if name == "" {
				name = args[0]
			}
			if proj == "" {
				proj = args[1]
			}
		case 1:
			if name == "" {
				name = args[0]
			}
		}
		if name == "" {
			return fmt.Errorf("agent name is required (first argument or --name)")
		}
		if proj == "" {
			return fmt.Errorf("project skills path is required (second argument or --project-path), e.g. .cursor/skills")
		}
		global := addAgentGlobalPath
		if global == "" {
			global = config.DeriveDefaultGlobalPath(proj)
		}
		if global == "" {
			return fmt.Errorf("could not derive global path from %q; set --global-path explicitly", proj)
		}

		okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.AddAgent(name, config.AgentConfig{
			ProjectPath: proj,
			GlobalPath:  global,
		})
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println(okStyle.Render(fmt.Sprintf("✓ Agent %q added", name)))
		fmt.Println(dimStyle.Render("  project_path: "+proj))
		fmt.Println(dimStyle.Render("  global_path:  "+global))
		fmt.Println(dimStyle.Render("  "+config.ConfigPath()))
		return nil
	},
}

// config remove-agent
var configRemoveAgentCmd = &cobra.Command{
	Use:   "remove-agent",
	Short: "Remove an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if !cfg.RemoveAgent(args[0]) {
			return fmt.Errorf("agent %q not found", args[0])
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(
			fmt.Sprintf("✓ Agent %q removed", args[0])))
		return nil
	},
}

// config add-project
var addProjectAgents []string

var configAddProjectCmd = &cobra.Command{
	Use:   "add-project <path>",
	Short: "Register a project directory for sync",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.AddProject(args[0], addProjectAgents)
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(
			fmt.Sprintf("✓ Project %q registered", args[0])))
		if len(addProjectAgents) > 0 {
			fmt.Println(dimStyle.Render(fmt.Sprintf("  Agents: %s", strings.Join(addProjectAgents, ", "))))
		} else {
			fmt.Println(dimStyle.Render("  Agents: auto-detect"))
		}
		return nil
	},
}

// config remove-project
var configRemoveProjectCmd = &cobra.Command{
	Use:   "remove-project <path>",
	Short: "Remove a registered project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if !cfg.RemoveProject(args[0]) {
			return fmt.Errorf("project %q not found", args[0])
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(
			fmt.Sprintf("✓ Project %q removed", args[0])))
		return nil
	},
}

// config list-projects
var configListProjectsCmd = &cobra.Command{
	Use:   "list-projects",
	Short: "List registered projects and their detected agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.Projects) == 0 {
			fmt.Println(dimStyle.Render("No projects registered."))
			fmt.Println(dimStyle.Render("Add with: skills-cli config add-project <path>"))
			return nil
		}

		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Bold(true)
		agentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

		fmt.Println()
		fmt.Println(titleStyle.Render(fmt.Sprintf("  %d registered project(s)", len(cfg.Projects))))
		fmt.Println()

		for _, proj := range cfg.Projects {
			fmt.Println(nameStyle.Render("  " + agent.ShortenPath(proj.Path)))

			agents := proj.Agents
			if len(agents) == 0 {
				agents = agent.DetectProjectAgents(proj.Path, cfg.Agents)
				if len(agents) > 0 {
					fmt.Printf("    %s %s\n", dimStyle.Render("detected:"), agentStyle.Render(strings.Join(agents, ", ")))
				} else {
					fmt.Println(dimStyle.Render("    no agents detected"))
				}
			} else {
				fmt.Printf("    %s %s\n", dimStyle.Render("configured:"), agentStyle.Render(strings.Join(agents, ", ")))
			}
		}
		fmt.Println()
		return nil
	},
}

// config set-sync
var setSyncAgents []string

var configSetSyncCmd = &cobra.Command{
	Use:   "set-sync",
	Short: "Set default sync agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(setSyncAgents) == 0 {
			return fmt.Errorf("--agents is required")
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.Sync.DefaultAgents = setSyncAgents
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(
			fmt.Sprintf("✓ Default sync agents set to: %s", strings.Join(setSyncAgents, ", "))))
		return nil
	},
}

// config set-watch-direction
var configSetWatchDirCmd = &cobra.Command{
	Use:   "set-watch-direction <direction>",
	Short: "Set watch direction (push|pull|both)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := config.NormalizeWatchDirection(args[0])
		valid := config.ValidWatchDirections()
		found := false
		for _, v := range valid {
			if v == dir {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid direction %q, must be one of: %s", args[0], strings.Join(valid, ", "))
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.Sync.Watch.Direction = dir
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(
			fmt.Sprintf("✓ Watch direction set to: %s", dir)))
		return nil
	},
}

// config list-agents
var configListAgentsCmd = &cobra.Command{
	Use:   "list-agents",
	Short: "List all configured agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Bold(true)

		var names []string
		for n := range cfg.Agents {
			names = append(names, n)
		}
		sort.Strings(names)

		fmt.Println()
		fmt.Println(titleStyle.Render(fmt.Sprintf("  %d configured agent(s)", len(names))))
		fmt.Println()
		for _, n := range names {
			ag := cfg.Agents[n]
			fmt.Printf("  %s\n", nameStyle.Render(n))
			fmt.Printf("    project: %s\n", dimStyle.Render(ag.ProjectPath))
			fmt.Printf("    global:  %s\n", dimStyle.Render(ag.GlobalPath))
		}
		fmt.Println()
		return nil
	},
}

func init() {
	configAddAgentCmd.Flags().StringVar(&addAgentName, "name", "", "Agent name")
	configAddAgentCmd.Flags().StringVar(&addAgentProjectPath, "project-path", "", "Project-level skills path")
	configAddAgentCmd.Flags().StringVar(&addAgentGlobalPath, "global-path", "", "Global skills path")

	configAddProjectCmd.Flags().StringSliceVar(&addProjectAgents, "agents", nil, "Specific agents for this project")

	configSetSyncCmd.Flags().StringSliceVar(&setSyncAgents, "agents", nil, "Default agent(s) for sync")

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configAddAgentCmd)
	configCmd.AddCommand(configRemoveAgentCmd)
	configCmd.AddCommand(configAddProjectCmd)
	configCmd.AddCommand(configRemoveProjectCmd)
	configCmd.AddCommand(configListProjectsCmd)
	configCmd.AddCommand(configListAgentsCmd)
	configCmd.AddCommand(configSetSyncCmd)
	configCmd.AddCommand(configSetWatchDirCmd)
}
