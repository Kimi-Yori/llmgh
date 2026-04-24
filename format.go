package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func tsv(fields ...string) {
	fmt.Println(strings.Join(fields, "\t"))
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
