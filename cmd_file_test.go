package main

import "testing"

func TestBuildContentsAPIPathEscapesPathSegments(t *testing.T) {
	got := buildContentsAPIPath("owner", "repo", "docs/file name#v1?.md", "feature/foo bar")
	want := "/repos/owner/repo/contents/docs/file%20name%23v1%3F.md?ref=feature%2Ffoo+bar"
	if got != want {
		t.Fatalf("buildContentsAPIPath() = %q, want %q", got, want)
	}
}

func TestBuildContentsAPIPathEscapesLiteralPercent(t *testing.T) {
	got := buildContentsAPIPath("owner", "repo", "docs/100% done.md", "")
	want := "/repos/owner/repo/contents/docs/100%25%20done.md"
	if got != want {
		t.Fatalf("buildContentsAPIPath() = %q, want %q", got, want)
	}
}
