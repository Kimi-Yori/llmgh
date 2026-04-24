package main

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

var tsvEscaper = strings.NewReplacer(
	"\\", "\\\\",
	"\t", "\\t",
	"\n", "\\n",
	"\r", "\\r",
)

var denseImageTagPattern = regexp.MustCompile(`!\[([^\]]*)\]\(https://www\.gstatic\.com/codereviewagent/[^)]*\)`)

func tsv(fields ...string) {
	escaped := make([]string, 0, len(fields))
	for _, field := range fields {
		escaped = append(escaped, formatField(field))
	}
	fmt.Println(strings.Join(escaped, "\t"))
}

func formatField(field string) string {
	if denseMode {
		field = compactDenseField(field)
	}
	field = tsvEscaper.Replace(field)
	if denseMode {
		field = normalizeDenseEscaped(field)
	}
	return field
}

func emitLegend(kind string) {
	if !denseMode {
		return
	}
	switch kind {
	case "pr":
		tsv("legend", "v=1", "merg=mergeable", "cmt=comment", "rc=review_comment", "chk=check", "err=error", "trunc=truncated")
	case "issue":
		tsv("legend", "v=1", "cmt=comment", "err=error", "trunc=truncated")
	}
}

func recordKind(full, dense string) string {
	if denseMode {
		return dense
	}
	return full
}

func str(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == math.Trunc(val) {
			return fmt.Sprintf("%.0f", val)
		}
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", val)
	}
}

func fmtTime(v any) string {
	s, ok := v.(string)
	if !ok || s == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	if denseMode {
		return t.Format("2006-01-02T15:04Z")
	}
	return t.Format("2006-01-02T15:04:05Z")
}

func fmtLabels(v any) string {
	arr, ok := v.([]any)
	if !ok {
		return ""
	}
	var names []string
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := m["name"].(string); ok {
			names = append(names, name)
		}
	}
	return strings.Join(names, ",")
}

func nested(data map[string]any, keys ...string) string {
	var current any = data
	for _, key := range keys {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = m[key]
	}
	return str(current)
}

func parseLimit(args []string, defaultVal int) (int, []string) {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--limit" {
			var n int
			fmt.Sscanf(args[i+1], "%d", &n)
			if n <= 0 {
				n = defaultVal
			}
			return n, append(args[:i], args[i+2:]...)
		}
	}
	return defaultVal, args
}

func parseFlag(args []string, flag string, defaultVal string) (string, []string) {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			val := args[i+1]
			return val, append(args[:i], args[i+2:]...)
		}
	}
	return defaultVal, args
}

func parseIntFlag(args []string, flag string, defaultVal int) (int, []string) {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			var n int
			fmt.Sscanf(args[i+1], "%d", &n)
			if n <= 0 {
				n = defaultVal
			}
			return n, append(args[:i], args[i+2:]...)
		}
	}
	return defaultVal, args
}

func parseSwitchFlag(args []string, flag string) (bool, []string) {
	for i, arg := range args {
		if arg == flag {
			return true, append(args[:i], args[i+1:]...)
		}
	}
	return false, args
}

func truncateText(s string, max int) string {
	if denseMode {
		s = sanitizeDenseText(s)
	}
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func sanitizeDenseText(s string) string {
	return denseImageTagPattern.ReplaceAllString(s, "[$1]")
}

func normalizeDenseEscaped(s string) string {
	for strings.Contains(s, `\n\n`) {
		s = strings.ReplaceAll(s, `\n\n`, `\n`)
	}
	return s
}

func compactDenseField(field string) string {
	switch {
	case field == "draft=false":
		return "draft=F"
	case field == "draft=true":
		return "draft=T"
	case field == "truncated=false":
		return "trunc=F"
	case field == "truncated=true":
		return "trunc=T"
	case field == "has_more=true":
		return "more=T"
	case strings.HasPrefix(field, "changed_files="):
		return "changed=" + strings.TrimPrefix(field, "changed_files=")
	case strings.HasPrefix(field, "private="):
		if strings.HasSuffix(field, "true") {
			return "private=T"
		}
		return "private=F"
	default:
		return field
	}
}
