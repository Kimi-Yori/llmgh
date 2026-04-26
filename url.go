package main

import (
	"fmt"
	"net/url"
	"strings"
)

const fileURLRefHint = "blob/tree URL parsing treats the first segment after blob/tree as the ref; for refs containing '/', use llmgh file get <path> --repo owner/repo --ref <ref>"

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
	case "file":
		if len(commandArgs) < 2 {
			return fmt.Errorf("url parse produced empty file command")
		}
		switch commandArgs[0] {
		case "get":
			if err := cmdFileGet(commandArgs[1:]); err != nil {
				return addFileURLHint(err)
			}
			return nil
		default:
			return fmt.Errorf("unsupported file command from url: %s", commandArgs[0])
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
	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host != "github.com" {
		return "", nil, fmt.Errorf("unsupported github url: %s", raw)
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
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
		if strings.HasPrefix(parsed.Fragment, "pullrequestreview-") {
			reviewID := strings.TrimPrefix(parsed.Fragment, "pullrequestreview-")
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
	case "blob", "tree":
		if len(parts) < 5 {
			return "", nil, fmt.Errorf("unsupported github url: %s", raw)
		}
		ref := parts[3]
		filePath := strings.Join(parts[4:], "/")
		return "file", []string{"get", filePath, "--repo", repoArg, "--ref", ref}, nil
	default:
		return "", nil, fmt.Errorf("unsupported github url: %s", raw)
	}
}

func addFileURLHint(err error) error {
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Kind != "not_found" {
		return err
	}

	withHint := *apiErr
	withHint.Message = withHint.Message + "; " + fileURLRefHint
	return &withHint
}
