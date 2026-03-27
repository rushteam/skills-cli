package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type TreeEntry struct {
	Path string `json:"path"`
	SHA  string `json:"sha"`
	Type string `json:"type"`
}

type TreeResponse struct {
	SHA  string      `json:"sha"`
	Tree []TreeEntry `json:"tree"`
}

func getGitHubToken() string {
	for _, key := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if token := os.Getenv(key); token != "" {
			return token
		}
	}
	return ""
}

func FetchSkillFolderHash(ownerRepo string, skillPath string, token string) (string, error) {
	if token == "" {
		token = getGitHubToken()
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/git/trees/main?recursive=1", ownerRepo)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var tree TreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return "", fmt.Errorf("failed to decode GitHub tree response: %w", err)
	}

	skillPath = strings.TrimSuffix(skillPath, "/SKILL.md")
	skillPath = strings.TrimSuffix(skillPath, "/")

	for _, entry := range tree.Tree {
		if entry.Type == "tree" && entry.Path == skillPath {
			return entry.SHA, nil
		}
	}

	return "", fmt.Errorf("skill path %q not found in repository tree", skillPath)
}
