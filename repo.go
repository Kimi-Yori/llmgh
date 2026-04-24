package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func resolveRepo(args []string) (owner, repo string, remaining []string, err error) {
	remaining = args
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--repo" {
			parts := strings.SplitN(args[i+1], "/", 2)
			if len(parts) != 2 {
				return "", "", nil, fmt.Errorf("--repo must be owner/repo format")
			}
			owner, repo = parts[0], parts[1]
			remaining = append(args[:i], args[i+2:]...)
			return
		}
	}

	owner, repo, err = detectFromGit()
	return
}

func detectFromGit() (string, string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", "", fmt.Errorf("no --repo and not in a git repository")
	}
	gitDir := strings.TrimSpace(string(out))
	out, err = exec.Command("git", "-C", gitDir, "remote", "get-url", "origin").Output()
	if err != nil {
		return "", "", fmt.Errorf("no --repo and git remote detection failed")
	}
	url := strings.TrimSpace(string(out))
	return parseGitURL(url)
}

func parseGitURL(url string) (string, string, error) {
	url = strings.TrimSuffix(url, ".git")

	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		parts := strings.Split(url, "/")
		if len(parts) >= 5 {
			return parts[len(parts)-2], parts[len(parts)-1], nil
		}
	}

	if strings.Contains(url, ":") {
		idx := strings.LastIndex(url, ":")
		path := url[idx+1:]
		parts := strings.SplitN(path, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	}

	return "", "", fmt.Errorf("cannot parse git remote URL: %s", url)
}
