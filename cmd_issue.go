package main

import (
	"fmt"
)

func cmdIssueView(args []string) error {
	owner, repo, args, err := resolveRepo(args)
	if err != nil {
		return err
	}
	if len(args) < 1 {
		return fmt.Errorf("usage: llmgh issue view <number>")
	}
	number := args[0]

	client, err := NewClient()
	if err != nil {
		return err
	}

	data, err := client.Get(fmt.Sprintf("/repos/%s/%s/issues/%s", owner, repo, number))
	if err != nil {
		return err
	}

	tsv("issue", str(data["number"]), str(data["state"]),
		nested(data, "user", "login"),
		str(data["title"]),
		fmtTime(data["created_at"]),
	)

	if labels := fmtLabels(data["labels"]); labels != "" {
		tsv("labels", labels)
	}

	if body := str(data["body"]); body != "" {
		if len(body) > 500 {
			body = body[:500] + "..."
		}
		tsv("body", body)
	}

	tsv("meta", "comments="+str(data["comments"]))

	return nil
}

func cmdIssueList(args []string) error {
	owner, repo, args, err := resolveRepo(args)
	if err != nil {
		return err
	}

	limit, args := parseLimit(args, 30)
	state, _ := parseFlag(args, "--state", "open")

	client, err := NewClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/repos/%s/%s/issues?state=%s&per_page=%d&sort=updated&direction=desc",
		owner, repo, state, limit)

	items, err := client.GetList(path)
	if err != nil {
		return err
	}

	for _, issue := range items {
		if issue["pull_request"] != nil {
			continue
		}
		tsv("issue", str(issue["number"]), str(issue["state"]),
			nested(issue, "user", "login"),
			str(issue["title"]),
			fmtLabels(issue["labels"]),
			fmtTime(issue["updated_at"]),
		)
	}

	if len(items) >= limit {
		tsv("page", "issues", fmt.Sprintf("shown=%d", len(items)), "has_more=true")
	}

	return nil
}
