package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"
)

type mergedComment struct {
	Kind      string
	ID        string
	User      string
	Path      string
	CreatedAt string
	Body      string
	SortTime  time.Time
}

func emitPRViewRecords(data map[string]any) {
	tsv("pr", str(data["number"]), str(data["state"]),
		nested(data, "base", "ref"), nested(data, "head", "ref"),
		nested(data, "user", "login"),
		str(data["title"]),
		fmtTime(data["created_at"]),
	)

	tsv(recordKind("mergeable", "merg"), str(data["mergeable"]),
		"draft="+str(data["draft"]),
		"comments="+str(data["comments"]),
		"changed_files="+str(data["changed_files"]),
		"+"+str(data["additions"]),
		"-"+str(data["deletions"]),
	)

	if body := str(data["body"]); body != "" {
		tsv("body", sanitizeText(body))
	}

	if labels := fmtLabels(data["labels"]); labels != "" {
		tsv("labels", labels)
	}
}

func emitIssueViewRecords(data map[string]any) {
	tsv("issue", str(data["number"]), str(data["state"]),
		nested(data, "user", "login"),
		str(data["title"]),
		fmtTime(data["created_at"]),
	)

	if labels := fmtLabels(data["labels"]); labels != "" {
		tsv("labels", labels)
	}

	if body := str(data["body"]); body != "" {
		tsv("body", sanitizeText(body))
	}

	tsv("meta", "comments="+str(data["comments"]))
}

func shortFileStatus(status string) string {
	switch status {
	case "added":
		return "A"
	case "removed":
		return "D"
	case "renamed":
		return "R"
	default:
		return "M"
	}
}

func emitSectionError(section string, err error) {
	if apiErr, ok := err.(*APIError); ok {
		tsv(recordKind("error", "err"), section,
			"kind="+apiErr.Kind,
			fmt.Sprintf("status=%d", apiErr.Status),
			apiErr.Message,
		)
		return
	}
	tsv(recordKind("error", "err"), section, err.Error())
}

func fetchPRReviews(client *Client, owner, repo, number string, limit int) ([]map[string]any, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%s/reviews?per_page=%d", owner, repo, number, limit)
	return client.GetList(path)
}

func fetchIssueComments(client *Client, owner, repo, number string, limit int) ([]map[string]any, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%s/comments?per_page=%d", owner, repo, number, limit)
	return client.GetList(path)
}

func fetchPRReviewComments(client *Client, owner, repo, number string, limit int) ([]map[string]any, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%s/comments?per_page=%d", owner, repo, number, limit)
	return client.GetList(path)
}

func mergePRComments(issueComments, reviewComments []map[string]any) []mergedComment {
	merged := make([]mergedComment, 0, len(issueComments)+len(reviewComments))

	for _, c := range issueComments {
		createdAt := fmtTime(c["created_at"])
		merged = append(merged, mergedComment{
			Kind:      "comment",
			ID:        str(c["id"]),
			User:      nested(c, "user", "login"),
			CreatedAt: createdAt,
			Body:      sanitizeText(str(c["body"])),
			SortTime:  parseSortTime(createdAt),
		})
	}

	for _, c := range reviewComments {
		createdAt := fmtTime(c["created_at"])
		merged = append(merged, mergedComment{
			Kind:      "review_comment",
			ID:        str(c["id"]),
			User:      nested(c, "user", "login"),
			Path:      str(c["path"]),
			CreatedAt: createdAt,
			Body:      sanitizeText(str(c["body"])),
			SortTime:  parseSortTime(createdAt),
		})
	}

	sort.SliceStable(merged, func(i, j int) bool {
		return merged[i].SortTime.After(merged[j].SortTime)
	})

	return merged
}

func sortIssueCommentsDesc(items []map[string]any) []mergedComment {
	comments := make([]mergedComment, 0, len(items))
	for _, c := range items {
		createdAt := fmtTime(c["created_at"])
		comments = append(comments, mergedComment{
			Kind:      "comment",
			ID:        str(c["id"]),
			User:      nested(c, "user", "login"),
			CreatedAt: createdAt,
			Body:      sanitizeText(str(c["body"])),
			SortTime:  parseSortTime(createdAt),
		})
	}

	sort.SliceStable(comments, func(i, j int) bool {
		return comments[i].SortTime.After(comments[j].SortTime)
	})

	return comments
}

func parseSortTime(value string) time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05Z", "2006-01-02T15:04Z"} {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

func intFromAny(v any) int {
	switch val := v.(type) {
	case float64:
		if val == math.Trunc(val) {
			return int(val)
		}
	case int:
		return val
	case string:
		n, err := strconv.Atoi(val)
		if err == nil {
			return n
		}
	}
	return 0
}
