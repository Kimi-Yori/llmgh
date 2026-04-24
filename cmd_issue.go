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

	emitIssueViewRecords(data)
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

func cmdIssueSummary(args []string) error {
	owner, repo, args, err := resolveRepo(args)
	if err != nil {
		return err
	}
	if len(args) < 1 {
		return fmt.Errorf("usage: llmgh issue summary <number>")
	}
	number := args[0]

	client, err := NewClient()
	if err != nil {
		return err
	}

	issueData, err := client.Get(fmt.Sprintf("/repos/%s/%s/issues/%s", owner, repo, number))
	if err != nil {
		return err
	}

	emitLegend("issue")
	emitIssueViewRecords(issueData)

	comments, err := fetchIssueComments(client, owner, repo, number, 100)
	if err != nil {
		emitSectionError("comments", err)
		return nil
	}

	sortedComments := sortIssueCommentsDesc(comments)
	shown := len(sortedComments)
	if shown > 10 {
		shown = 10
	}

	totalComments := intFromAny(issueData["comments"])
	tsv(recordKind("comments_meta", "comments_meta"),
		"issue="+number,
		fmt.Sprintf("total=%d", totalComments),
		fmt.Sprintf("shown=%d", shown),
		fmt.Sprintf("truncated=%t", totalComments > shown),
	)

	for _, comment := range sortedComments[:shown] {
		tsv(recordKind("comment", "cmt"), number, comment.ID, comment.User, comment.CreatedAt, comment.Body)
	}

	return nil
}
