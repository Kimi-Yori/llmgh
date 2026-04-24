package main

import (
	"fmt"
	"strings"
)

func cmdURL(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: llmgh url <github-url>")
	}

	command, commandArgs, err := parseGitHubURL(args[0])
	if err != nil {
		return err
	}

	switch command {
	case "status":
		return cmdStatus(commandArgs)
	case "pr":
		if len(commandArgs) < 1 {
			return fmt.Errorf("url parse produced empty pr command")
		}
		switch commandArgs[0] {
		case "view":
			return cmdPRView(commandArgs[1:])
		case "files":
			return cmdPRFiles(commandArgs[1:])
		case "review-detail":
			return cmdPRReviewDetail(commandArgs[1:])
		default:
			return fmt.Errorf("unsupported pr command from url: %s", commandArgs[0])
		}
	case "issue":
		if len(commandArgs) < 1 {
			return fmt.Errorf("url parse produced empty issue command")
		}
		switch commandArgs[0] {
		case "view":
			return cmdIssueView(commandArgs[1:])
		default:
			return fmt.Errorf("unsupported issue command from url: %s", commandArgs[0])
		}
	default:
		return fmt.Errorf("unsupported command from url: %s", command)
	}
}

func parseGitHubURL(raw string) (string, []string, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "https://")
	trimmed = strings.TrimPrefix(trimmed, "http://")
	if !strings.HasPrefix(trimmed, "github.com/") {
		return "", nil, fmt.Errorf("unsupported github url: %s", raw)
	}

	pathAndFragment := strings.TrimPrefix(trimmed, "github.com/")
	path, fragment, _ := strings.Cut(pathAndFragment, "#")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return "", nil, fmt.Errorf("unsupported github url: %s", raw)
	}

	owner := parts[0]
	repo := parts[1]
	repoArg := owner + "/" + repo

	if len(parts) == 2 {
		return "status", []string{"--repo", repoArg}, nil
	}

	switch parts[2] {
	case "pull":
		if len(parts) < 4 {
			return "", nil, fmt.Errorf("unsupported github url: %s", raw)
		}
		number := parts[3]
		if strings.HasPrefix(fragment, "pullrequestreview-") {
			reviewID := strings.TrimPrefix(fragment, "pullrequestreview-")
			if reviewID == "" {
				return "", nil, fmt.Errorf("unsupported github url: %s", raw)
			}
			return "pr", []string{"review-detail", number, reviewID, "--repo", repoArg}, nil
		}
		if len(parts) >= 5 && parts[4] == "files" {
			return "pr", []string{"files", number, "--repo", repoArg}, nil
		}
		return "pr", []string{"view", number, "--repo", repoArg}, nil
	case "issues":
		if len(parts) < 4 {
			return "", nil, fmt.Errorf("unsupported github url: %s", raw)
		}
		return "issue", []string{"view", parts[3], "--repo", repoArg}, nil
	default:
		return "", nil, fmt.Errorf("unsupported github url: %s", raw)
	}
}
