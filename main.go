package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"text/tabwriter"
)

func main() {
	forge := "github"
	err := run(forge)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", forge, err)
		os.Exit(1)
	}
}

func run(forge string) error {
	result, err := GithubCacheOrFetch()
	if err != nil {
		return fmt.Errorf("failed to get github PRs: %w", err)
	}

	if err := writeCache(forge, result); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write cache (but will continue): %+v", err)
	}

	cmd := exec.Command("rofi", "-width", "70", "-dmenu", "-theme", "Arc-Dark", "-i")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin for rofi: %w", err)
	}
	var out strings.Builder
	cmd.Stdout = &out

	defer func() {
		_ = stdin.Close()
	}()

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("failed to start rofi subprocess: %w", err)
	}

	tw := tabwriter.NewWriter(stdin, 2, 2, 2, ' ', 0)

	nodes := []githubResultNode{}
	nodes = append(nodes, result.Data.Requests.Nodes...)
	nodes = append(nodes, result.Data.Created.Nodes...)
	nodes = append(nodes, result.Data.Mentions.Nodes...)
	nodes = append(nodes, result.Data.Assigned.Nodes...)
	for _, res := range nodes {
		// TODO: Reorder to this order:
		// fmt.Fprintf(tw, "%s\t%s\t#%d\t%s\n", res.Repository.NameWithOwner, res.HeadRef.Name, res.Number, res.Title)
		_, err := fmt.Fprintf(tw, "%s\t%d\t%s\t%s\n", res.Repository.NameWithOwner, res.Number, res.HeadRef.Name, res.Title)
		if err != nil {
			return fmt.Errorf("failed to write to rofi pipe: %w", err)
		}
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("failed to flush pipe to rofi: %w", err)
	}
	if _, err := stdin.Write([]byte("refresh\n")); err != nil {
		return fmt.Errorf("failed to write to rofi pipe: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("rofi failed: %w", err)
	}

	res := out.String()
	if res == "refresh" {
		// TODO
		os.Exit(0)
	}
	re := regexp.MustCompile(`([^/]+)/([^[:space:]]+)[[:space:]]+([0-9]+)[[:space:]]+([^[:space:]]+)`)
	matched := re.FindStringSubmatch(res)
	if matched == nil {
		// They aborted rofi, by for example pressing ESC
		return nil
	}
	owner := matched[1]
	repo := matched[2]
	pr := matched[3]
	// branch := matched[4]

	cmd = exec.Command("xdg-open", fmt.Sprintf("https://github.com/%s/%s/pull/%s", owner, repo, pr))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed top run xdg-open: %w", err)
	}

	return nil
}

func pullCache(forge string) string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find cache directory, falling back to ~/.cache/: %s\n", err)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to find home directory, giving up: %s\n", err)
			os.Exit(1)
		}
		cacheDir = path.Join(homeDir, ".cache")
	}
	return path.Join(cacheDir, "gitforge-rofi", fmt.Sprintf("%s-pulls.json", forge))
}

func writeCache(forge string, data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := os.MkdirAll(path.Dir(pullCache(forge)), 0750); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}

	if err := os.WriteFile(pullCache(forge), b, 0644); err != nil {
		return fmt.Errorf("failed to write to cache: %w", err)
	}
	return nil
}

func configDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find config directory, falling back to ~/.config/: %s\n", err)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to find home directory, giving up: %s\n", err)
			os.Exit(1)
		}
		return path.Join(homeDir, ".config", "gitforge-rofi")
	}
	return path.Join(configDir, "gitforge-rofi")
}
