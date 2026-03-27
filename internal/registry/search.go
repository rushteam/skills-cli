package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const DefaultAPIBase = "https://skills.sh"

type SearchResult struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Source   string `json:"source"`
	Installs int    `json:"installs"`
}

type searchResponse struct {
	Skills []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Installs int    `json:"installs"`
		Source   string `json:"source"`
	} `json:"skills"`
}

func SearchSkills(query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	apiURL := fmt.Sprintf("%s/api/search?q=%s&limit=%d", DefaultAPIBase, url.QueryEscape(query), limit)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("search API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned status %d", resp.StatusCode)
	}

	var data searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	var results []SearchResult
	for _, s := range data.Skills {
		results = append(results, SearchResult{
			Name:     s.Name,
			Slug:     s.ID,
			Source:   s.Source,
			Installs: s.Installs,
		})
	}
	return results, nil
}

func FormatInstalls(count int) string {
	if count <= 0 {
		return ""
	}
	if count >= 1_000_000 {
		return fmt.Sprintf("%.1fM installs", float64(count)/1_000_000)
	}
	if count >= 1_000 {
		return fmt.Sprintf("%.1fK installs", float64(count)/1_000)
	}
	if count == 1 {
		return "1 install"
	}
	return fmt.Sprintf("%d installs", count)
}
