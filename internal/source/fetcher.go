package source

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rushteam/skills-cli/internal/skill"
)

func FetchSkills(ps ParsedSource) (string, []*skill.Skill, error) {
	if ps.Type == SourceLocal {
		searchPath := ps.LocalPath
		if ps.Subpath != "" {
			searchPath = filepath.Join(searchPath, ps.Subpath)
		}
		skills, err := skill.DiscoverSkills(searchPath)
		if err != nil {
			return ps.LocalPath, nil, err
		}
		return ps.LocalPath, skills, nil
	}

	tmpDir, err := os.MkdirTemp("", "skills-cli-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	cloneOpts := &git.CloneOptions{
		URL:   ps.URL,
		Depth: 1,
	}
	if ps.Ref != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(ps.Ref)
		cloneOpts.SingleBranch = true
	}

	_, err = git.PlainClone(tmpDir, false, cloneOpts)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", nil, fmt.Errorf("git clone failed: %w", err)
	}

	searchPath := tmpDir
	if ps.Subpath != "" {
		searchPath = filepath.Join(tmpDir, ps.Subpath)
	}

	skills, err := skill.DiscoverSkills(searchPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", nil, err
	}

	if ps.SkillFilter != "" {
		var filtered []*skill.Skill
		for _, s := range skills {
			if s.Name == ps.SkillFilter {
				filtered = append(filtered, s)
			}
		}
		skills = filtered
	}

	return tmpDir, skills, nil
}

func Cleanup(tmpDir string, ps ParsedSource) {
	if ps.Type != SourceLocal && tmpDir != "" {
		os.RemoveAll(tmpDir)
	}
}
