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
