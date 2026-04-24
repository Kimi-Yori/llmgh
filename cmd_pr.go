package main

import (
	"fmt"
)

func cmdPRView(args []string) error {
	owner, repo, args, err := resolveRepo(args)
	if err != nil {
		return err
	}
	if len(args) < 1 {
		return fmt.Errorf("usage: llmgh pr view <number>")
	}
	number := args[0]

	client, err := NewClient()
	if err != nil {
		return err
	}

	data, err := client.Get(fmt.Sprintf("/repos/%s/%s/pulls/%s", owner, repo, number))
	if err != nil {
		return err
	}

	tsv("pr", str(data["number"]), str(data["state"]),
		nested(data, "base", "ref"), nested(data, "head", "ref"),
		nested(data, "user", "login"),
		str(data["title"]),
		fmtTime(data["created_at"]),
	)

	tsv("mergeable", str(data["mergeable"]),
		"draft="+str(data["draft"]),
		"comments="+str(data["comments"]),
		"changed_files="+str(data["changed_files"]),
		"+"+str(data["additions"]),
		"-"+str(data["deletions"]),
	)

	if body := str(data["body"]); body != "" {
		if len(body) > 500 {
			body = body[:500] + "..."
		}
		tsv("body", body)
	}

	if labels := fmtLabels(data["labels"]); labels != "" {
		tsv("labels", labels)
	}

	return nil
}

func cmdPRList(args []string) error {
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

	path := fmt.Sprintf("/repos/%s/%s/pulls?state=%s&per_page=%d&sort=updated&direction=desc",
		owner, repo, state, limit)

	items, err := client.GetList(path)
	if err != nil {
		return err
	}

	for _, pr := range items {
		tsv("pr", str(pr["number"]), str(pr["state"]),
			nested(pr, "base", "ref"), nested(pr, "head", "ref"),
			nested(pr, "user", "login"),
			str(pr["title"]),
			fmtTime(pr["updated_at"]),
		)
	}

	if len(items) >= limit {
		tsv("page", "prs", fmt.Sprintf("shown=%d", len(items)), "has_more=true")
	}

	return nil
}

func cmdPRFiles(args []string) error {
	owner, repo, args, err := resolveRepo(args)
	if err != nil {
		return err
	}
	if len(args) < 1 {
		return fmt.Errorf("usage: llmgh pr files <number>")
	}
	number := args[0]
	limit, _ := parseLimit(args[1:], 100)

	client, err := NewClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/repos/%s/%s/pulls/%s/files?per_page=%d",
		owner, repo, number, limit)

	items, err := client.GetList(path)
	if err != nil {
		return err
	}

	for _, f := range items {
		status := str(f["status"])
		short := "M"
		switch status {
		case "added":
			short = "A"
		case "removed":
			short = "D"
		case "renamed":
			short = "R"
		}
		tsv("file", short, str(f["filename"]),
			"+"+str(f["additions"]), "-"+str(f["deletions"]),
		)
	}

	if len(items) >= limit {
		tsv("page", "files", fmt.Sprintf("shown=%d", len(items)), "has_more=true")
	}

	return nil
}

func cmdPRChecks(args []string) error {
	owner, repo, args, err := resolveRepo(args)
	if err != nil {
		return err
	}
	if len(args) < 1 {
		return fmt.Errorf("usage: llmgh pr checks <number>")
	}
	number := args[0]

	client, err := NewClient()
	if err != nil {
		return err
	}

	prData, err := client.Get(fmt.Sprintf("/repos/%s/%s/pulls/%s", owner, repo, number))
	if err != nil {
		return err
	}

	sha := nested(prData, "head", "sha")
	if sha == "" {
		return fmt.Errorf("cannot determine head SHA")
	}

	data, err := client.Get(fmt.Sprintf("/repos/%s/%s/commits/%s/check-runs?per_page=100", owner, repo, sha))
	if err != nil {
		return err
	}

	runs, ok := data["check_runs"].([]any)
	if !ok {
		tsv("checks", "none")
		return nil
	}

	for _, run := range runs {
		r, ok := run.(map[string]any)
		if !ok {
			continue
		}
		conclusion := str(r["conclusion"])
		if conclusion == "" {
			conclusion = str(r["status"])
		}
		tsv("check", str(r["name"]), conclusion)
	}

	return nil
}

func cmdPRComments(args []string) error {
	owner, repo, args, err := resolveRepo(args)
	if err != nil {
		return err
	}
	if len(args) < 1 {
		return fmt.Errorf("usage: llmgh pr comments <number>")
	}
	number := args[0]
	limit, _ := parseLimit(args[1:], 30)

	client, err := NewClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/repos/%s/%s/issues/%s/comments?per_page=%d",
		owner, repo, number, limit)
	comments, err := client.GetList(path)
	if err != nil {
		return err
	}

	for _, c := range comments {
		body := str(c["body"])
		if len(body) > 300 {
			body = body[:300] + "..."
		}
		tsv("comment", str(c["id"]),
			nested(c, "user", "login"),
			fmtTime(c["created_at"]),
			body,
		)
	}

	path = fmt.Sprintf("/repos/%s/%s/pulls/%s/comments?per_page=%d",
		owner, repo, number, limit)
	reviews, err := client.GetList(path)
	if err != nil {
		return err
	}

	for _, c := range reviews {
		body := str(c["body"])
		if len(body) > 300 {
			body = body[:300] + "..."
		}
		tsv("review_comment", str(c["id"]),
			nested(c, "user", "login"),
			str(c["path"]),
			fmtTime(c["created_at"]),
			body,
		)
	}

	total := len(comments) + len(reviews)
	if total == 0 {
		tsv("comments", "none")
	}

	return nil
}
