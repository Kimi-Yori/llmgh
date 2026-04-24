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

	emitPRViewRecords(data)
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
		tsv("file", shortFileStatus(str(f["status"])), str(f["filename"]),
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

	runs, err := fetchPRCheckRuns(client, owner, repo, nested(prData, "head", "sha"))
	if err != nil {
		return err
	}

	if len(runs) == 0 {
		tsv("checks", "none")
		return nil
	}

	for _, run := range runs {
		conclusion := str(run["conclusion"])
		if conclusion == "" {
			conclusion = str(run["status"])
		}
		tsv(recordKind("check", "chk"), str(run["name"]), conclusion)
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

	comments, err := fetchIssueComments(client, owner, repo, number, limit)
	if err != nil {
		return err
	}

	for _, c := range comments {
		tsv(recordKind("comment", "cmt"), str(c["id"]),
			nested(c, "user", "login"),
			fmtTime(c["created_at"]),
			truncateText(str(c["body"]), 300),
		)
	}

	reviews, err := fetchPRReviewComments(client, owner, repo, number, limit)
	if err != nil {
		return err
	}

	for _, c := range reviews {
		tsv(recordKind("review_comment", "rc"), str(c["id"]),
			nested(c, "user", "login"),
			str(c["path"]),
			fmtTime(c["created_at"]),
			truncateText(str(c["body"]), 300),
		)
	}

	total := len(comments) + len(reviews)
	if total == 0 {
		tsv("comments", "none")
	}

	return nil
}

func cmdPRReviews(args []string) error {
	owner, repo, args, err := resolveRepo(args)
	if err != nil {
		return err
	}
	if len(args) < 1 {
		return fmt.Errorf("usage: llmgh pr reviews <number>")
	}
	number := args[0]
	limit, _ := parseLimit(args[1:], 30)

	client, err := NewClient()
	if err != nil {
		return err
	}

	reviews, err := fetchPRReviews(client, owner, repo, number, limit)
	if err != nil {
		return err
	}

	for _, review := range reviews {
		tsv(recordKind("review", "rv"), str(review["id"]),
			nested(review, "user", "login"),
			str(review["state"]),
			fmtTime(review["created_at"]),
			truncateText(str(review["body"]), 300),
		)
	}

	if len(reviews) >= limit {
		tsv("page", "reviews", fmt.Sprintf("shown=%d", len(reviews)), "has_more=true")
	}

	return nil
}

func cmdPRReviewDetail(args []string) error {
	owner, repo, args, err := resolveRepo(args)
	if err != nil {
		return err
	}
	if len(args) < 2 {
		return fmt.Errorf("usage: llmgh pr review-detail <number> <review_id>")
	}
	number := args[0]
	reviewID := args[1]

	client, err := NewClient()
	if err != nil {
		return err
	}

	review, err := client.Get(fmt.Sprintf("/repos/%s/%s/pulls/%s/reviews/%s", owner, repo, number, reviewID))
	if err != nil {
		return err
	}

	tsv("review", str(review["id"]),
		nested(review, "user", "login"),
		str(review["state"]),
		fmtTime(review["created_at"]),
		truncateText(str(review["body"]), 300),
	)

	comments, err := client.GetList(fmt.Sprintf("/repos/%s/%s/pulls/%s/reviews/%s/comments?per_page=100", owner, repo, number, reviewID))
	if err != nil {
		return err
	}

	for _, comment := range comments {
		tsv(recordKind("review_comment", "rc"), reviewID,
			str(comment["id"]),
			nested(comment, "user", "login"),
			str(comment["path"]),
			fmtTime(comment["created_at"]),
			truncateText(str(comment["body"]), 300),
		)
	}

	return nil
}

func cmdPRSummary(args []string) error {
	owner, repo, args, err := resolveRepo(args)
	if err != nil {
		return err
	}
	if len(args) < 1 {
		return fmt.Errorf("usage: llmgh pr summary <number>")
	}
	number := args[0]
	maxFiles, args := parseIntFlag(args[1:], "--max-files", 15)
	maxComments, args := parseIntFlag(args, "--max-comments", 10)
	includeChecks, _ := parseSwitchFlag(args, "--checks")

	client, err := NewClient()
	if err != nil {
		return err
	}

	prData, err := client.Get(fmt.Sprintf("/repos/%s/%s/pulls/%s", owner, repo, number))
	if err != nil {
		return err
	}

	emitPRViewRecords(prData)

	apiWait()
	files, err := client.GetList(fmt.Sprintf("/repos/%s/%s/pulls/%s/files?per_page=%d", owner, repo, number, maxFiles))
	if err != nil {
		emitSectionError("files", err)
	} else {
		totalFiles := intFromAny(prData["changed_files"])
		tsv(recordKind("files_meta", "files_meta"),
			"pr="+number,
			fmt.Sprintf("total=%d", totalFiles),
			fmt.Sprintf("shown=%d", len(files)),
			fmt.Sprintf("truncated=%t", totalFiles > len(files)),
		)
		for _, f := range files {
			tsv("file", number, shortFileStatus(str(f["status"])), str(f["filename"]),
				"+"+str(f["additions"]), "-"+str(f["deletions"]),
			)
		}
	}

	if includeChecks {
		apiWait()
		runs, err := fetchPRCheckRuns(client, owner, repo, nested(prData, "head", "sha"))
		if err != nil {
			emitSectionError("checks", err)
		} else if len(runs) == 0 {
			tsv("checks", number, "none")
		} else {
			for _, run := range runs {
				conclusion := str(run["conclusion"])
				if conclusion == "" {
					conclusion = str(run["status"])
				}
				tsv(recordKind("check", "chk"), number, str(run["name"]), conclusion)
			}
		}
	}

	apiWait()
	reviews, err := fetchPRReviews(client, owner, repo, number, 100)
	if err != nil {
		emitSectionError("reviews", err)
	} else if len(reviews) == 0 {
		tsv("reviews", number, "none")
	} else {
		for _, review := range reviews {
			tsv("review", number, str(review["id"]),
				nested(review, "user", "login"),
				str(review["state"]),
				fmtTime(review["created_at"]),
				truncateText(str(review["body"]), 300),
			)
		}
	}

	apiWait()
	issueComments, issueErr := fetchIssueComments(client, owner, repo, number, 100)
	if issueErr != nil {
		emitSectionError("comments_issue", issueErr)
	}

	apiWait()
	reviewComments, reviewErr := fetchPRReviewComments(client, owner, repo, number, 100)
	if reviewErr != nil {
		emitSectionError("comments_review", reviewErr)
	}

	if issueErr == nil || reviewErr == nil {
		merged := mergePRComments(issueComments, reviewComments)
		shown := len(merged)
		if shown > maxComments {
			shown = maxComments
		}

		if issueErr == nil && reviewErr == nil {
			totalComments := intFromAny(prData["comments"]) + intFromAny(prData["review_comments"])
			tsv(recordKind("comments_meta", "comments_meta"),
				"pr="+number,
				fmt.Sprintf("total=%d", totalComments),
				fmt.Sprintf("shown=%d", shown),
				fmt.Sprintf("truncated=%t", totalComments > shown),
			)
		} else {
			tsv(recordKind("comments_meta", "comments_meta"),
				"pr="+number,
				"total=unknown",
				fmt.Sprintf("shown=%d", shown),
				"partial=true",
			)
		}

		for _, comment := range merged[:shown] {
			switch comment.Kind {
			case "review_comment":
				tsv(recordKind("review_comment", "rc"), number, comment.ID, comment.User, comment.Path, comment.CreatedAt, comment.Body)
			default:
				tsv(recordKind("comment", "cmt"), number, comment.ID, comment.User, comment.CreatedAt, comment.Body)
			}
		}
	}

	return nil
}

func fetchPRCheckRuns(client *Client, owner, repo, sha string) ([]map[string]any, error) {
	if sha == "" {
		return nil, fmt.Errorf("cannot determine head SHA")
	}

	data, err := client.Get(fmt.Sprintf("/repos/%s/%s/commits/%s/check-runs?per_page=100", owner, repo, sha))
	if err != nil {
		return nil, err
	}

	rawRuns, ok := data["check_runs"].([]any)
	if !ok {
		return nil, nil
	}

	runs := make([]map[string]any, 0, len(rawRuns))
	for _, run := range rawRuns {
		r, ok := run.(map[string]any)
		if !ok {
			continue
		}
		runs = append(runs, r)
	}

	return runs, nil
}
