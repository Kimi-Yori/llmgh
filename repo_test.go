package main

import "testing"

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
	}{
		{
			name:      "https with dot git",
			url:       "https://github.com/octocat/hello-world.git",
			wantOwner: "octocat",
			wantRepo:  "hello-world",
		},
		{
			name:      "https without dot git",
			url:       "https://github.com/octocat/hello-world",
			wantOwner: "octocat",
			wantRepo:  "hello-world",
		},
		{
			name:      "ssh with dot git",
			url:       "git@github.com:octocat/hello-world.git",
			wantOwner: "octocat",
			wantRepo:  "hello-world",
		},
		{
			name:      "ssh without dot git",
			url:       "git@github.com:octocat/hello-world",
			wantOwner: "octocat",
			wantRepo:  "hello-world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOwner, gotRepo, err := parseGitURL(tt.url)
			if err != nil {
				t.Fatalf("parseGitURL(%q) returned error: %v", tt.url, err)
			}
			if gotOwner != tt.wantOwner || gotRepo != tt.wantRepo {
				t.Fatalf("parseGitURL(%q) = (%q, %q), want (%q, %q)",
					tt.url, gotOwner, gotRepo, tt.wantOwner, tt.wantRepo)
			}
		})
	}
}
