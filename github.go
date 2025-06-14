package main

import (
	"bytes"
	"encoding/base64"
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

type githubResult struct {
	Data githubResultData `json:"data"`
}

type githubResultData struct {
	Requests githubResultNodes `json:"requests"`
	Created  githubResultNodes `json:"created"`
	Mentions githubResultNodes `json:"mentions"`
	Assigned githubResultNodes `json:"assigned"`
}
type githubResultNodes struct {
	Nodes []githubResultNode `json:"nodes"`
}
type githubResultNode struct {
	Number     uint
	Url        string
	State      string // OPEN | ?
	Title      string
	CreatedAt  string // ISO-8601
	Repository githubResultRepository
	HeadRef    githubResultHeadRef
}

type githubResultRepository struct {
	NameWithOwner string
}
type githubResultHeadRef struct {
	Name string
}

func graphQlSearch(name string, query string) string {
	return fmt.Sprintf(`
  %s: search(query: "is:open is:pr %s archived:false", type: ISSUE, first: 100) {
    nodes {
      ... on PullRequest {
        number
        url
        state
        title
        createdAt
        repository {
          nameWithOwner
        }
        headRef {
          name
        }
      }
    }
  }
`, name, query)
}

type GithubForge struct{}

func (f GithubForge) name() string {
	return "github"
}

func (f GithubForge) list() error {
	res, err := f.cachedFetch()
	if err != nil {
		return fmt.Errorf("listing github PRs failed: %w", err)
	}

	nodes := []githubResultNode{}
	nodes = append(nodes, res.Data.Requests.Nodes...)
	nodes = append(nodes, res.Data.Created.Nodes...)
	nodes = append(nodes, res.Data.Mentions.Nodes...)
	nodes = append(nodes, res.Data.Assigned.Nodes...)
	for _, node := range nodes {
		pr := fmt.Sprintf("%s#%d", node.Repository.NameWithOwner, node.Number)
		fmt.Printf("%-40s %-50s [%s]", pr, node.Title, node.HeadRef.Name)
		fmt.Printf("\000info\x1f%s", node.Url)
		fmt.Print("\x1fmeta\x1fgithub")
		fmt.Println()
	}

	return nil
}

func (f GithubForge) refresh() error {
	res, err := GithubFetch()
	if err != nil {
		return fmt.Errorf("failed to fetch github PRs: %w", err)
	}
	if err := writeCache(f.name(), res); err != nil {
		return fmt.Errorf("failed to write github cache: %w", err)
	}

	return nil
}

func (f GithubForge) cachedFetch() (*githubResult, error) {
	rawCache, err := readCache(f.name())
	if err != nil {
		return nil, fmt.Errorf("failed to read %s cache: %w", f.name(), err)
	}

	if len(rawCache) != 0 {
		var result githubResult
		if err := json.Unmarshal(rawCache, &result); err != nil {
			return nil, fmt.Errorf("failed to parse %s cache: %w", f.name(), err)
		}
		return &result, nil
	}

	res, err := GithubFetch()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch github PRs: %w", err)
	}
	if err := writeCache(f.name(), res); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write github cache (but will continue): %v", err)
	}
	return res, nil
}

func GithubFetch() (*githubResult, error) {
	forge := "github"
	tokenPath := path.Join(configDir(), forge, "token")
	rawToken, err := os.ReadFile(tokenPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("no token found for forge %s in %q", forge, tokenPath)
		}
		return nil, fmt.Errorf("failed to read token for forge %s: %w", forge, err)
	}

	userPath := path.Join(configDir(), forge, "username")
	rawUser, err := os.ReadFile(userPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("no username found for forge %s in %q", forge, userPath)
		}
		return nil, fmt.Errorf("failed to read username for forge %s: %w", forge, err)
	}
	user := strings.TrimSpace(string(rawUser))

	query := "{\n" + strings.Join([]string{
		graphQlSearch("requests", "review-requested:"+user),
		graphQlSearch("created", "author:"+user),
		graphQlSearch("mentions", "mentions:"+user),
		graphQlSearch("assigned", "assigned:"+user),
	}, "\n") + "\n}"
	data := struct {
		Query string `json:"query"`
	}{
		Query: query,
	}
	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize JSON data: %w", err)
	}
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(serialized))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}
	token := strings.TrimSpace(string(rawToken))
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString(fmt.Appendf([]byte{}, "%s:%s", user, token)))
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

	var result githubResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &result, nil
}
