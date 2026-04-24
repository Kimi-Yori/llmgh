# llmgh

Compact GitHub context reader for LLM agents.

llmgh is not a replacement for `gh`.
Default output is stable TSV, optimized for token efficiency.

## Install

```bash
CGO_ENABLED=0 go build -ldflags "-s -w" -o llmgh .
cp llmgh ~/tools/llmgh/
```

## Auth

Token resolution order:

1. `LLMGH_TOKEN`
2. `GH_TOKEN`
3. `GITHUB_TOKEN`
4. `gh auth token` (fallback, requires gh CLI)

## Commands

### status

```bash
llmgh status --repo owner/repo
```

```
repo	owner/repo	default_branch=main	private=true	stars=0	forks=0
auth	ok	user=alice
rate	core	remaining=4998	limit=5000	reset=1777000000
```

### pr view

```bash
llmgh pr view 123 --repo owner/repo
```

```
pr	123	open	main	feature/foo	alice	Fix auth flow	2026-04-23T10:12:00Z
mergeable	true	draft=false	comments=3	changed_files=5	+120	-30
body	Summary of changes...
labels	bug,security
```

### pr list

```bash
llmgh pr list --repo owner/repo --state open --limit 10
```

```
pr	123	open	main	feature/foo	alice	Fix auth flow	2026-04-24T10:00:00Z
pr	120	open	main	fix/bar	bob	Fix rendering	2026-04-23T15:00:00Z
page	prs	shown=10	has_more=true
```

### pr files

```bash
llmgh pr files 123 --repo owner/repo
```

```
file	M	src/auth.go	+32	-8
file	A	src/auth_test.go	+45	-0
file	D	src/old_auth.go	+0	-120
```

### pr checks

```bash
llmgh pr checks 123 --repo owner/repo
```

```
check	ci/build	success
check	ci/lint	success
check	ci/test	failure
```

### pr comments

```bash
llmgh pr comments 123 --repo owner/repo
```

```
comment	9876	bob	2026-04-23T11:00:00Z	Looks good except token handling.
review_comment	5432	alice	src/auth.go	2026-04-23T12:00:00Z	Should we validate the token here?
```

### issue view

```bash
llmgh issue view 456 --repo owner/repo
```

```
issue	456	open	alice	Bug in auth flow	2026-04-20T08:00:00Z
labels	bug,ClaudeCode
body	Description of the issue...
meta	comments=5
```

### issue list

```bash
llmgh issue list --repo owner/repo --state open --limit 10
```

```
issue	456	open	alice	Bug in auth flow	bug,ClaudeCode	2026-04-24T10:00:00Z
issue	450	open	bob	Add retry logic	enhancement	2026-04-22T14:00:00Z
```

## Output format

- TSV (tab-separated)
- 1 line per record
- Column 1 is always the record kind (`pr`, `issue`, `file`, `check`, `comment`, `review_comment`, `page`, `meta`, `repo`, `auth`, `rate`)
- Errors go to stderr in `ERR\tkind\tmessage` format
- No ANSI colors

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Usage error |
| 2 | Auth error |
| 3 | Not found |
| 4 | Rate limited |
| 5 | Network error |
| 6 | API error |

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `--repo owner/repo` | Target repository | Detect from git remote |
| `--limit N` | Max results | 30 |
| `--state open\|closed\|all` | Filter state | open |

## Design

- REST API v3 only (no GraphQL dependency)
- Standard library only (`net/http` + `encoding/json`)
- Static binary, zero external dependencies
- Token efficiency: 60-90% fewer tokens than `gh` JSON output
- `gh` CLI is optional (token fallback only)

## Comparison

```
# gh pr view (GraphQL, verbose)
Title:  Fix auth flow
State:  OPEN
Author: alice
Labels: bug, security
Assignees: bob
...25+ lines

# llmgh pr view (REST, compact)
pr	123	open	main	feature/foo	alice	Fix auth flow	2026-04-23T10:12:00Z
mergeable	true	draft=false	comments=3	changed_files=5	+120	-30
labels	bug,security
...4 lines
```

## License

MIT
