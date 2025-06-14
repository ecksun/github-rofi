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
	"time"
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

func GithubCacheOrFetch() (*githubResult, error) {
	forge := "github"

	cacheFile := pullCache(forge)
	fileinfo, err := os.Stat(cacheFile)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("failed to check file age of %q: %w", cacheFile, err)
	}
	if err == nil && fileinfo.ModTime().Add(180*time.Minute).After(time.Now()) {
		rawCache, err := os.ReadFile(cacheFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read github pull cache %q: %w", cacheFile, err)
		}
		var result githubResult
		if err := json.Unmarshal(rawCache, &result); err != nil {
			return nil, fmt.Errorf("failed to parse cache file %q: %w", cacheFile, err)
		}
		fmt.Fprintf(os.Stderr, "Read data from cache successfully\n")
		return &result, nil
	}
	return GithubFetch()
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
