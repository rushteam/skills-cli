package agent

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rushteam/skills-cli/internal/config"
)

type SkillInfo struct {
	Name      string
	Path      string
	Agent     string
	Scope     string // "central", "global", "project"
	Project   string // project path if scope is "project"
}

func DetectInstalledAgents(cfg *config.Config) []string {
	var installed []string
	for name, agent := range cfg.Agents {
		globalDir := config.ResolveGlobalPath(agent)
		parent := filepath.Dir(globalDir)
		if dirExists(parent) {
			installed = append(installed, name)
		}
	}
	sort.Strings(installed)
	return installed
}

func DetectProjectAgents(projectPath string, agents map[string]config.AgentConfig) []string {
	var detected []string
	seen := make(map[string]bool)
	for name, agent := range agents {
		agentDir := filepath.Join(projectPath, agent.ProjectPath)
		parentDir := filepath.Dir(agentDir)
		if !seen[name] && dirExists(parentDir) {
			detected = append(detected, name)
			seen[name] = true
		}
	}
	sort.Strings(detected)
	return detected
}

func ResolveProjectSkillsDir(projectPath string, agentName string, agents map[string]config.AgentConfig) string {
	agent, ok := agents[agentName]
	if !ok {
		return ""
	}
	return filepath.Join(projectPath, agent.ProjectPath)
}

func ScanSkillsInDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var skills []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillMd := filepath.Join(dir, entry.Name(), "SKILL.md")
		if fileExists(skillMd) {
			skills = append(skills, entry.Name())
		}
	}
	sort.Strings(skills)
	return skills, nil
}

func ListCentralSkills() ([]SkillInfo, error) {
	dir := config.SkillsHome()
	names, err := ScanSkillsInDir(dir)
	if err != nil {
		return nil, err
	}
	var result []SkillInfo
	for _, name := range names {
		result = append(result, SkillInfo{
			Name:  name,
			Path:  filepath.Join(dir, name),
			Scope: "central",
		})
	}
	return result, nil
}

func ListGlobalSkills(cfg *config.Config) ([]SkillInfo, error) {
	var result []SkillInfo
	seen := make(map[string]bool)
	for agentName, agent := range cfg.Agents {
		dir := config.ResolveGlobalPath(agent)
		key := dir
		if seen[key] {
			continue
		}
		seen[key] = true
		names, err := ScanSkillsInDir(dir)
		if err != nil {
			continue
		}
		for _, name := range names {
			result = append(result, SkillInfo{
				Name:  name,
				Path:  filepath.Join(dir, name),
				Agent: agentName,
				Scope: "global",
			})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Agent != result[j].Agent {
			return result[i].Agent < result[j].Agent
		}
		return result[i].Name < result[j].Name
	})
	return result, nil
}

func ListProjectSkills(cfg *config.Config) ([]SkillInfo, error) {
	var result []SkillInfo
	for _, proj := range cfg.Projects {
		agents := proj.Agents
		if len(agents) == 0 {
			agents = DetectProjectAgents(proj.Path, cfg.Agents)
		}
		for _, agentName := range agents {
			dir := ResolveProjectSkillsDir(proj.Path, agentName, cfg.Agents)
			if dir == "" {
				continue
			}
			names, err := ScanSkillsInDir(dir)
			if err != nil {
				continue
			}
			for _, name := range names {
				result = append(result, SkillInfo{
					Name:    name,
					Path:    filepath.Join(dir, name),
					Agent:   agentName,
					Scope:   "project",
					Project: proj.Path,
				})
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Project != result[j].Project {
			return result[i].Project < result[j].Project
		}
		if result[i].Agent != result[j].Agent {
			return result[i].Agent < result[j].Agent
		}
		return result[i].Name < result[j].Name
	})
	return result, nil
}

func ListAllSkills(cfg *config.Config) ([]SkillInfo, error) {
	var all []SkillInfo

	central, err := ListCentralSkills()
	if err == nil {
		all = append(all, central...)
	}

	global, err := ListGlobalSkills(cfg)
	if err == nil {
		all = append(all, global...)
	}

	project, err := ListProjectSkills(cfg)
	if err == nil {
		all = append(all, project...)
	}

	return all, nil
}

func ShortenPath(fullPath string) string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(fullPath, home) {
		return "~" + fullPath[len(home):]
	}
	return fullPath
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
