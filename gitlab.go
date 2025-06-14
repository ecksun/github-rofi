package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

type GitlabForge struct{}

func (f GitlabForge) name() string {
	return "gitlab"
}
func (f GitlabForge) list() error {
	res, err := gitlabCacheOrFetch()
	if err != nil {
		return fmt.Errorf("listing gitlab MRs failed: %w", err)
	}

	for _, mr := range res {
		fmt.Printf("%-40s %-50s [%s]", mr.References.Full, mr.Title, mr.Branch)
		fmt.Printf("\000info\x1f%s", mr.Url)
		fmt.Print("\x1fmeta\x1fgitlab")
		fmt.Println()
	}
	return nil
}
func (f GitlabForge) refresh() error {
	res, err := fetch()
	if err != nil {
		return fmt.Errorf("failed to fetch gitlab MRs: %w", err)
	}

	if err := writeCache(f.name(), res); err != nil {
		return fmt.Errorf("failed to write gitlab cache: %w", err)
	}

	return nil
}

func gitlabCacheOrFetch() ([]gitlabMR, error) {
	forge := "gitlab"

	rawCache, err := readCache(forge)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s cache: %w", forge, err)
	}

	if len(rawCache) != 0 {
		var result []gitlabMR
		if err := json.Unmarshal(rawCache, &result); err != nil {
			return nil, fmt.Errorf("failed to parse %q cache: %w", forge, err)
		}
		return result, nil
	}

	res, err := fetch()
	if err != nil {
		return res, fmt.Errorf("failed to fetch gitlab MRs: %w", err)
	}

	if err := writeCache(forge, res); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write gitlab cache (but will continue): %v", err)
	}
	return res, err
}

type gitlabMR struct {
	Title      string             `json:"title"`
	Url        string             `json:"web_url"`
	CreatedAt  string             `json:"created_at"`
	References gitlabMRReferences `json:"references"`
	Branch     string             `json:"source_branch"`
}

type gitlabMRReferences struct {
	Full string `json:"full"`
}

func fetch() ([]gitlabMR, error) {
	forge := "gitlab"
	tokenPath := path.Join(configDir(), forge, "token")
	rawToken, err := os.ReadFile(tokenPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("no token found for forge %s in %q", forge, tokenPath)
		}
		return nil, fmt.Errorf("failed to read token for forge %s: %w", forge, err)
	}

	token := strings.TrimSpace(string(rawToken))

	created, err := listMrs(token, "created_by_me")
	if err != nil {
		return nil, fmt.Errorf("failed to list MRs created by me: %w", err)
	}
	assigned, err := listMrs(token, "assigned_to_me")
	if err != nil {
		return nil, fmt.Errorf("failed to list MRs assigned to me: %w", err)
	}

	mrs := []gitlabMR{}
	mrs = append(mrs, created...)
	mrs = append(mrs, assigned...)

	return mrs, nil
}

func listMrs(token string, scope string) ([]gitlabMR, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://gitlab.com/api/v4/merge_requests?state=opened&scope=%s", scope), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create MR GET request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do HTTP request: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result []gitlabMR
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return result, nil
}
