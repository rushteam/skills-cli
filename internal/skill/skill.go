package skill

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Skill struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Path        string `yaml:"-" json:"path"`
	RawContent  string `yaml:"-" json:"-"`
}

func ParseSkillMd(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	frontmatter, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	var skill Skill
	if err := yaml.Unmarshal([]byte(frontmatter), &skill); err != nil {
		return nil, err
	}
	if skill.Name == "" || skill.Description == "" {
		return nil, nil
	}
	skill.Path = filepath.Dir(path)
	skill.RawContent = content
	return &skill, nil
}

func extractFrontmatter(content string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var lines []string
	inFrontmatter := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break
		}
		if inFrontmatter {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n"), nil
}

func DiscoverSkills(basePath string) ([]*Skill, error) {
	var skills []*Skill
	seen := make(map[string]bool)

	skillMdPath := filepath.Join(basePath, "SKILL.md")
	if fileExists(skillMdPath) {
		if s, err := ParseSkillMd(skillMdPath); err == nil && s != nil && !seen[s.Name] {
			skills = append(skills, s)
			seen[s.Name] = true
		}
	}

	searchDirs := []string{
		basePath,
		filepath.Join(basePath, "skills"),
		filepath.Join(basePath, "skills", ".curated"),
		filepath.Join(basePath, "skills", ".experimental"),
		filepath.Join(basePath, "skills", ".system"),
		filepath.Join(basePath, ".agents", "skills"),
		filepath.Join(basePath, ".claude", "skills"),
		filepath.Join(basePath, ".cursor", "skills"),
		filepath.Join(basePath, ".codebuddy", "skills"),
		filepath.Join(basePath, ".codex", "skills"),
		filepath.Join(basePath, ".trae", "skills"),
		filepath.Join(basePath, ".windsurf", "skills"),
		filepath.Join(basePath, ".qoder", "skills"),
		filepath.Join(basePath, ".roo", "skills"),
		filepath.Join(basePath, ".goose", "skills"),
		filepath.Join(basePath, ".continue", "skills"),
		filepath.Join(basePath, ".openhands", "skills"),
	}

	for _, dir := range searchDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			mdPath := filepath.Join(dir, entry.Name(), "SKILL.md")
			if !fileExists(mdPath) {
				continue
			}
			s, err := ParseSkillMd(mdPath)
			if err != nil || s == nil || seen[s.Name] {
				continue
			}
			skills = append(skills, s)
			seen[s.Name] = true
		}
	}

	return skills, nil
}

func ListSkillFiles(skillDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(skillDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, _ := filepath.Rel(skillDir, path)
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
