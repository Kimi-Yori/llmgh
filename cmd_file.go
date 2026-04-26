package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func cmdFileGet(args []string) error {
	owner, repo, args, err := resolveRepo(args)
	if err != nil {
		return err
	}

	ref, args := parseFlag(args, "--ref", "")

	if len(args) < 1 {
		return fmt.Errorf("usage: llmgh file get <path> [--repo owner/repo] [--ref branch]")
	}
	path := args[0]

	client, err := NewClient()
	if err != nil {
		return err
	}

	apiPath := buildContentsAPIPath(owner, repo, path, ref)

	body, err := client.GetRaw(apiPath)
	if err != nil {
		return err
	}

	if _, err := os.Stdout.Write(body); err != nil {
		return fmt.Errorf("write stdout: %w", err)
	}
	return nil
}

func buildContentsAPIPath(owner, repo, filePath, ref string) string {
	apiPath := fmt.Sprintf("/repos/%s/%s/contents/%s",
		url.PathEscape(owner),
		url.PathEscape(repo),
		escapePathSegments(filePath),
	)
	if ref == "" {
		return apiPath
	}

	query := url.Values{}
	query.Set("ref", ref)
	return apiPath + "?" + query.Encode()
}

func escapePathSegments(path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}
