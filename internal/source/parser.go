package source

import (
	"path/filepath"
	"regexp"
	"strings"
)

type SourceType string

const (
	SourceGitHub    SourceType = "github"
	SourceGitLab    SourceType = "gitlab"
	SourceLocal     SourceType = "local"
	SourceGit       SourceType = "git"
	SourceWellKnown SourceType = "well-known"
)

type ParsedSource struct {
	Type        SourceType
	URL         string
	Ref         string
	Subpath     string
	SkillFilter string
	LocalPath   string
}

var (
	githubTreePathRe = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/tree/([^/]+)/(.+)`)
	githubTreeRe     = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/tree/([^/]+)$`)
	githubRepoRe     = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)`)
	gitlabTreePathRe = regexp.MustCompile(`^(https?):\/\/([^/]+)\/(.+?)\/-\/tree\/([^/]+)\/(.+)`)
	gitlabTreeRe     = regexp.MustCompile(`^(https?):\/\/([^/]+)\/(.+?)\/-\/tree\/([^/]+)$`)
	gitlabRepoRe     = regexp.MustCompile(`gitlab\.com\/(.+?)(?:\.git)?\/?$`)
	atSkillRe        = regexp.MustCompile(`^([^/]+)/([^/@]+)@(.+)$`)
	shorthandRe      = regexp.MustCompile(`^([^/]+)/([^/]+)(?:/(.+))?$`)
	sshRe            = regexp.MustCompile(`^git@[^:]+:(.+)$`)
)

func Parse(input string) ParsedSource {
	if isLocalPath(input) {
		abs, err := filepath.Abs(input)
		if err != nil {
			abs = input
		}
		return ParsedSource{
			Type:      SourceLocal,
			URL:       abs,
			LocalPath: abs,
		}
	}

	if m := githubTreePathRe.FindStringSubmatch(input); m != nil {
		owner, repo, ref, subpath := m[1], m[2], m[3], m[4]
		return ParsedSource{
			Type:    SourceGitHub,
			URL:     "https://github.com/" + owner + "/" + repo + ".git",
			Ref:     ref,
			Subpath: subpath,
		}
	}

	if m := githubTreeRe.FindStringSubmatch(input); m != nil {
		owner, repo, ref := m[1], m[2], m[3]
		return ParsedSource{
			Type: SourceGitHub,
			URL:  "https://github.com/" + owner + "/" + repo + ".git",
			Ref:  ref,
		}
	}

	if m := githubRepoRe.FindStringSubmatch(input); m != nil {
		owner, repo := m[1], m[2]
		repo = strings.TrimSuffix(repo, ".git")
		return ParsedSource{
			Type: SourceGitHub,
			URL:  "https://github.com/" + owner + "/" + repo + ".git",
		}
	}

	if m := gitlabTreePathRe.FindStringSubmatch(input); m != nil {
		protocol, hostname, repoPath, ref, subpath := m[1], m[2], m[3], m[4], m[5]
		if hostname != "github.com" {
			return ParsedSource{
				Type:    SourceGitLab,
				URL:     protocol + "://" + hostname + "/" + strings.TrimSuffix(repoPath, ".git") + ".git",
				Ref:     ref,
				Subpath: subpath,
			}
		}
	}

	if m := gitlabTreeRe.FindStringSubmatch(input); m != nil {
		protocol, hostname, repoPath, ref := m[1], m[2], m[3], m[4]
		if hostname != "github.com" {
			return ParsedSource{
				Type: SourceGitLab,
				URL:  protocol + "://" + hostname + "/" + strings.TrimSuffix(repoPath, ".git") + ".git",
				Ref:  ref,
			}
		}
	}

	if m := gitlabRepoRe.FindStringSubmatch(input); m != nil {
		repoPath := m[1]
		if strings.Contains(repoPath, "/") {
			return ParsedSource{
				Type: SourceGitLab,
				URL:  "https://gitlab.com/" + repoPath + ".git",
			}
		}
	}

	if m := atSkillRe.FindStringSubmatch(input); m != nil && !strings.Contains(input, ":") && !strings.HasPrefix(input, ".") {
		owner, repo, skillFilter := m[1], m[2], m[3]
		return ParsedSource{
			Type:        SourceGitHub,
			URL:         "https://github.com/" + owner + "/" + repo + ".git",
			SkillFilter: skillFilter,
		}
	}

	if m := shorthandRe.FindStringSubmatch(input); m != nil && !strings.Contains(input, ":") && !strings.HasPrefix(input, ".") {
		owner, repo, subpath := m[1], m[2], m[3]
		return ParsedSource{
			Type:    SourceGitHub,
			URL:     "https://github.com/" + owner + "/" + repo + ".git",
			Subpath: subpath,
		}
	}

	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		if !strings.HasSuffix(input, ".git") {
			return ParsedSource{
				Type: SourceWellKnown,
				URL:  input,
			}
		}
	}

	return ParsedSource{
		Type: SourceGit,
		URL:  input,
	}
}

func GetOwnerRepo(ps ParsedSource) string {
	if ps.Type == SourceLocal {
		return ""
	}
	if m := sshRe.FindStringSubmatch(ps.URL); m != nil {
		path := strings.TrimSuffix(m[1], ".git")
		if strings.Contains(path, "/") {
			return path
		}
		return ""
	}
	if !strings.HasPrefix(ps.URL, "http://") && !strings.HasPrefix(ps.URL, "https://") {
		return ""
	}
	parts := strings.Split(strings.TrimPrefix(strings.TrimPrefix(ps.URL, "https://"), "http://"), "/")
	if len(parts) >= 3 {
		repo := strings.TrimSuffix(parts[2], ".git")
		return parts[1] + "/" + repo
	}
	return ""
}

func isLocalPath(input string) bool {
	if filepath.IsAbs(input) {
		return true
	}
	return strings.HasPrefix(input, "./") || strings.HasPrefix(input, "../") || input == "." || input == ".."
}
