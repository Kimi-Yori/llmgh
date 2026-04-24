package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "status":
		if err := cmdStatus(args); err != nil {
			exitError(err)
		}
	case "pr":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "ERR\tusage\tllmgh pr <view|list|files|checks|comments> [args]")
			os.Exit(1)
		}
		sub := args[0]
		subArgs := args[1:]
		switch sub {
		case "view":
			if err := cmdPRView(subArgs); err != nil {
				exitError(err)
			}
		case "list":
			if err := cmdPRList(subArgs); err != nil {
				exitError(err)
			}
		case "files":
			if err := cmdPRFiles(subArgs); err != nil {
				exitError(err)
			}
		case "checks":
			if err := cmdPRChecks(subArgs); err != nil {
				exitError(err)
			}
		case "comments":
			if err := cmdPRComments(subArgs); err != nil {
				exitError(err)
			}
		default:
			fmt.Fprintf(os.Stderr, "ERR\tusage\tunknown pr subcommand: %s\n", sub)
			os.Exit(1)
		}
	case "issue":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "ERR\tusage\tllmgh issue <view|list> [args]")
			os.Exit(1)
		}
		sub := args[0]
		subArgs := args[1:]
		switch sub {
		case "view":
			if err := cmdIssueView(subArgs); err != nil {
				exitError(err)
			}
		case "list":
			if err := cmdIssueList(subArgs); err != nil {
				exitError(err)
			}
		default:
			fmt.Fprintf(os.Stderr, "ERR\tusage\tunknown issue subcommand: %s\n", sub)
			os.Exit(1)
		}
	case "--version", "version":
		fmt.Printf("llmgh %s\n", version)
	case "--help", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "ERR\tusage\tunknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `llmgh %s - compact GitHub context reader for LLM agents

Usage: llmgh <command> [args]

Commands:
  status                 repo/auth/rate info
  pr view <number>       PR details
  pr list [--state X]    list PRs
  pr files <number>      changed files
  pr checks <number>     CI check status
  issue view <number>    issue details
  issue list [--state X] list issues

Options:
  --repo owner/repo      target repository (default: detect from git remote)
  --limit N              max results (default: 30)

Auth: LLMGH_TOKEN > GH_TOKEN > GITHUB_TOKEN
`, version)
}

func exitError(err error) {
	if apiErr, ok := err.(*APIError); ok {
		fmt.Fprintf(os.Stderr, "ERR\t%s\tstatus=%d\t%s\n", apiErr.Kind, apiErr.Status, apiErr.Message)
		os.Exit(apiErr.ExitCode)
	}
	fmt.Fprintf(os.Stderr, "ERR\tinternal\t%v\n", err)
	os.Exit(1)
}
