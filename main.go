package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
)

func main() {

	fmt.Fprintf(os.Stderr, "ROFI_RETV=%s\n", os.Getenv("ROFI_RETV"))

	switch os.Getenv("ROFI_RETV") {
	case "": // Called directly
		cmd := exec.Command("rofi", "-show", "fb", "-modes", "fb: "+os.Args[0], "-width", "70", "-theme", "Arc-Dark", "-i")
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start rofi subprocess: %+v", err)
			os.Exit(1)
		}
		if err := cmd.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "rofi failed: %+v", err)
			os.Exit(1)
		}
	case "0": // Rofi: Initial call of script.
		forge := "github"
		err := run(forge)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", forge, err)
			os.Exit(1)
		}
	case "1": // Rofi: Selected an entry.
		if os.Args[1] == "refresh" {
			if _, err := GithubFetch(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to refresh github pull requests: %+v", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
		cmd := exec.Command("xdg-open", os.Getenv("ROFI_INFO"))
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "failed top run xdg-open: %+v", err)
			os.Exit(1)
		}
	case "2": // Rofi: Selected a custom entry.
	case "3": // Rofi: Deleted an entry.
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

	fmt.Println("\000prompt\x1fGithub PR")

	nodes := []githubResultNode{}
	nodes = append(nodes, result.Data.Requests.Nodes...)
	nodes = append(nodes, result.Data.Created.Nodes...)
	nodes = append(nodes, result.Data.Mentions.Nodes...)
	nodes = append(nodes, result.Data.Assigned.Nodes...)
	for _, res := range nodes {
		pr := fmt.Sprintf("%s#%d", res.Repository.NameWithOwner, res.Number)
		fmt.Printf("%-40s %-50s [%s]", pr, res.Title, res.HeadRef.Name)
		fmt.Printf("\000info\x1f%s", res.Url)
		fmt.Print("\x1fmeta\x1fgithub")
		fmt.Println()
	}
	fmt.Println("refresh")

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
