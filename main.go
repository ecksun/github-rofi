package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"time"
)

func main() {
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
		forge := GithubForge{}
		err := run(forge)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", forge.name(), err)
			os.Exit(1)
		}
	case "1": // Rofi: Selected an entry.
		if os.Args[1] == "refresh" {
			forge := GithubForge{}
			if err := forge.refresh(); err != nil {
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

type forge interface {
	name() string
	list() error
	refresh() error
}

func run(forge forge) error {
	fmt.Println("\000prompt\x1fGitforge changesets")

	err := forge.list()
	if err != nil {
		fmt.Println("refresh")
		return fmt.Errorf("failed to get %s PRs: %w", forge.name(), err)
	}

	fmt.Println("refresh")

	return nil
}

func cachePath(forge string) string {
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
	return path.Join(cacheDir, "gitforge-rofi", fmt.Sprintf("%s-changesets.json", forge))
}

func writeCache(forge string, data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := os.MkdirAll(path.Dir(cachePath(forge)), 0750); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}

	if err := os.WriteFile(cachePath(forge), b, 0644); err != nil {
		return fmt.Errorf("failed to write to cache: %w", err)
	}
	return nil
}

func readCache(forge string) ([]byte, error) {
	cacheFile := cachePath(forge)
	fileinfo, err := os.Stat(cacheFile)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("failed to check file age of %q: %w", cacheFile, err)
	}
	if err == nil && fileinfo.ModTime().Add(180*time.Minute).After(time.Now()) {
		rawCache, err := os.ReadFile(cacheFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read github pull cache %q: %w", cacheFile, err)
		}
		return rawCache, nil
	}
	return []byte{}, nil
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
