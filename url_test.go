package main

import (
	"reflect"
	"testing"
)

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		wantCommand string
		wantArgs    []string
		wantErr     bool
	}{
		{
			name:        "pull view",
			raw:         "https://github.com/owner/repo/pull/123",
			wantCommand: "pr",
			wantArgs:    []string{"view", "123", "--repo", "owner/repo"},
		},
		{
			name:        "pull review detail",
			raw:         "https://github.com/owner/repo/pull/123#pullrequestreview-456",
			wantCommand: "pr",
			wantArgs:    []string{"review-detail", "123", "456", "--repo", "owner/repo"},
		},
		{
			name:        "pull files",
			raw:         "https://github.com/owner/repo/pull/123/files",
			wantCommand: "pr",
			wantArgs:    []string{"files", "123", "--repo", "owner/repo"},
		},
		{
			name:        "issue view",
			raw:         "https://github.com/owner/repo/issues/456",
			wantCommand: "issue",
			wantArgs:    []string{"view", "456", "--repo", "owner/repo"},
		},
		{
			name:        "repo status",
			raw:         "https://github.com/owner/repo",
			wantCommand: "status",
			wantArgs:    []string{"--repo", "owner/repo"},
		},
		{
			name:        "blob file",
			raw:         "https://github.com/mattpocock/skills/blob/main/improve-codebase-architecture/LANGUAGE.md",
			wantCommand: "file",
			wantArgs:    []string{"get", "improve-codebase-architecture/LANGUAGE.md", "--repo", "mattpocock/skills", "--ref", "main"},
		},
		{
			name:        "blob nested path",
			raw:         "https://github.com/owner/repo/blob/develop/src/lib/utils.ts",
			wantCommand: "file",
			wantArgs:    []string{"get", "src/lib/utils.ts", "--repo", "owner/repo", "--ref", "develop"},
		},
		{
			name:        "blob query and line fragment are ignored",
			raw:         "https://github.com/owner/repo/blob/main/src/lib/utils.ts?plain=1#L10",
			wantCommand: "file",
			wantArgs:    []string{"get", "src/lib/utils.ts", "--repo", "owner/repo", "--ref", "main"},
		},
		{
			name:        "blob encoded path is unescaped",
			raw:         "https://github.com/owner/repo/blob/main/docs/file%20name%23v1%3F.md",
			wantCommand: "file",
			wantArgs:    []string{"get", "docs/file name#v1?.md", "--repo", "owner/repo", "--ref", "main"},
		},
		{
			name:        "blob slash ref limitation",
			raw:         "https://github.com/owner/repo/blob/feature/foo/src/lib/utils.ts",
			wantCommand: "file",
			wantArgs:    []string{"get", "foo/src/lib/utils.ts", "--repo", "owner/repo", "--ref", "feature"},
		},
		{
			name:        "tree path",
			raw:         "https://github.com/owner/repo/tree/main/src/lib",
			wantCommand: "file",
			wantArgs:    []string{"get", "src/lib", "--repo", "owner/repo", "--ref", "main"},
		},
		{
			name:    "invalid url",
			raw:     "https://example.com/owner/repo/pull/123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCommand, gotArgs, err := parseGitHubURL(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseGitHubURL(%q) returned nil error", tt.raw)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseGitHubURL(%q) returned error: %v", tt.raw, err)
			}
			if gotCommand != tt.wantCommand {
				t.Fatalf("parseGitHubURL(%q) command = %q, want %q", tt.raw, gotCommand, tt.wantCommand)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Fatalf("parseGitHubURL(%q) args = %#v, want %#v", tt.raw, gotArgs, tt.wantArgs)
			}
		})
	}
}
