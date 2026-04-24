package main

import (
	"io"
	"os"
	"testing"
)

func setDenseMode(t *testing.T, mode bool) {
	t.Helper()
	prev := denseMode
	denseMode = mode
	t.Cleanup(func() {
		denseMode = prev
	})
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	os.Stdout = writer
	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close: %v", err)
	}
	os.Stdout = oldStdout

	out, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("reader.Close: %v", err)
	}

	return string(out)
}

func TestTSVEscapesSpecialCharacters(t *testing.T) {
	setDenseMode(t, true)

	out := captureStdout(t, func() {
		tsv("kind", "line1\nline2\tpath\\name")
	})

	want := "kind\tline1\\nline2\\tpath\\\\name\n"
	if out != want {
		t.Fatalf("tsv() = %q, want %q", out, want)
	}
}

func TestTSVDenseNormalizesDoubleNewlines(t *testing.T) {
	setDenseMode(t, true)

	out := captureStdout(t, func() {
		tsv("kind", "a\n\nb")
	})

	want := "kind\ta\\nb\n"
	if out != want {
		t.Fatalf("tsv() = %q, want %q", out, want)
	}
}

func TestTruncateText(t *testing.T) {
	setDenseMode(t, true)

	tests := []struct {
		name string
		in   string
		max  int
		want string
	}{
		{name: "empty", in: "", max: 5, want: ""},
		{name: "exact", in: "hello", max: 5, want: "hello"},
		{name: "over", in: "hello!", max: 5, want: "hello..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateText(tt.in, tt.max); got != tt.want {
				t.Fatalf("truncateText(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
			}
		})
	}
}

func TestTruncateTextDenseSanitizesGeminiTag(t *testing.T) {
	setDenseMode(t, true)

	in := "![high](https://www.gstatic.com/codereviewagent/high-priority.svg)"
	if got := truncateText(in, 100); got != "[high]" {
		t.Fatalf("truncateText(%q, 100) = %q, want %q", in, got, "[high]")
	}
}

func TestStr(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "nil", in: nil, want: ""},
		{name: "string", in: "hello", want: "hello"},
		{name: "float integer", in: float64(42), want: "42"},
		{name: "float decimal", in: float64(3.5), want: "3.5"},
		{name: "bool true", in: true, want: "true"},
		{name: "bool false", in: false, want: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := str(tt.in); got != tt.want {
				t.Fatalf("str(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDenseAndFullSwitch(t *testing.T) {
	t.Run("dense", func(t *testing.T) {
		setDenseMode(t, true)

		if got := recordKind("comment", "cmt"); got != "cmt" {
			t.Fatalf("recordKind dense = %q, want %q", got, "cmt")
		}
		if got := fmtTime("2026-04-20T09:13:53Z"); got != "2026-04-20T09:13Z" {
			t.Fatalf("fmtTime dense = %q, want %q", got, "2026-04-20T09:13Z")
		}
	})

	t.Run("full", func(t *testing.T) {
		setDenseMode(t, false)

		if got := recordKind("comment", "cmt"); got != "comment" {
			t.Fatalf("recordKind full = %q, want %q", got, "comment")
		}
		if got := fmtTime("2026-04-20T09:13:53Z"); got != "2026-04-20T09:13:53Z" {
			t.Fatalf("fmtTime full = %q, want %q", got, "2026-04-20T09:13:53Z")
		}
	})
}
