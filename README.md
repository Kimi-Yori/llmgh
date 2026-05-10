<table>
    <thead>
        <tr>
            <th style="text-align:center">English</th>
            <th style="text-align:center"><a href="README_ja.md">日本語</a></th>
        </tr>
    </thead>
</table>

# llmgh

Compact GitHub context reader for LLM agents.

llmgh is not a replacement for `gh`. It reads GitHub data (PRs, issues, files, checks) and outputs stable TSV optimized for token efficiency. Designed for Claude Code, Cursor, Codex, and any LLM coding agent.

## Why llmgh?

`gh pr view` outputs 25+ lines of verbose, human-formatted text. LLM agents pay for every token — both in the output and in the reasoning required to parse it.

llmgh gives 60-90% fewer tokens with zero information loss.

## Install

**Pre-built binary** (Linux x86_64):

```bash
curl -L -o llmgh https://github.com/Kimi-Yori/llmgh/releases/latest/download/llmgh-linux-amd64
chmod +x llmgh
sudo mv llmgh /usr/local/bin/
```

**Build from source** (requires Go 1.22+):

```bash
git clone https://github.com/Kimi-Yori/llmgh.git
cd llmgh
make build
sudo make install
```

**Cross-build:**

```bash
make release    # linux/amd64, darwin/amd64, darwin/arm64
# Output: dist/llmgh-{os}-{arch}
```

## Auth

Token resolution order:

1. `LLMGH_TOKEN`
2. `GH_TOKEN`
3. `GITHUB_TOKEN`
4. `gh auth token` (fallback, requires gh CLI)

## Commands

### url

```bash
llmgh url https://github.com/owner/repo/pull/123#pullrequestreview-456
```

Parses a GitHub URL and dispatches to the corresponding subcommand (`status` / `pr` / `issue` / `file`).

For `blob`/`tree` URLs, the first segment after `blob`/`tree` is treated as the ref. For refs containing `/` (e.g. `feature/foo`), use `file get` explicitly:

```bash
llmgh file get src/lib/utils.ts --repo owner/repo --ref feature/foo
```

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

### pr reviews

```bash
llmgh pr reviews 123 --repo owner/repo
```

```
review	1111	alice	APPROVED	2026-04-23T12:30:00Z	Looks good to me.
review	1112	bob	CHANGES_REQUESTED	2026-04-23T12:45:00Z	Please fix token validation.
```

### pr review-detail

```bash
llmgh pr review-detail 123 1112 --repo owner/repo
```

```
review	1112	bob	CHANGES_REQUESTED	2026-04-23T12:45:00Z	Please fix token validation.
review_comment	1112	5432	bob	src/auth.go	2026-04-23T12:46:00Z	Validate the token before storing it.
```

### pr summary

```bash
llmgh pr summary 123 --repo owner/repo
llmgh pr summary 123 --repo owner/repo --checks
llmgh pr summary 123 --repo owner/repo --max-files 5 --max-comments 3
```

Dense output (default):
```
legend	v=1	merg=mergeable	cmt=comment	rc=review_comment	chk=check	err=error	trunc=truncated
pr	123	open	main	feature/foo	alice	Fix auth flow	2026-04-23T10:12Z
merg		draft=F	comments=3	changed=5	+120	-30
labels	bug,security
files_meta	pr=123	total=5	shown=5	trunc=F
file	123	M	src/auth.go	+32	-8
review	123	1111	alice	APPROVED	2026-04-23T12:30Z	Looks good to me.
comments_meta	pr=123	total=4	shown=4	trunc=F
cmt	123	9876	bob	2026-04-23T11:00Z	Looks good except token handling.
rc	123	5432	alice	src/auth.go	2026-04-23T12:00Z	Should we validate the token here?
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

### issue summary

```bash
llmgh issue summary 456 --repo owner/repo
```

```
legend	v=1	cmt=comment	err=error	trunc=truncated
issue	456	open	alice	Bug in auth flow	2026-04-20T08:00Z
labels	bug,ClaudeCode
body	Description of the issue...
meta	comments=5
comments_meta	issue=456	total=5	shown=5	trunc=F
cmt	456	9999	bob	2026-04-24T10:00Z	Needs a retry path.
```

### file get

```bash
llmgh file get path/to/file --repo owner/repo --ref main
llmgh url https://github.com/owner/repo/blob/main/path/to/file
```

Fetches raw file content via GitHub Contents API. Path segments are URL-encoded, so filenames with spaces or `?`/`#` work correctly.

## Output Format

- TSV (tab-separated), dense by default (`--full` for verbose)
- 1 line per record
- Tabs/newlines/carriage returns in field values escaped as `\t`, `\n`, `\r`
- Summary commands emit a `legend` row as the first line (dense only)
- Dense record kinds: `merg`(mergeable), `cmt`(comment), `rc`(review_comment), `chk`(check), `err`(error)
- Full record kinds (`--full`): `mergeable`, `comment`, `review_comment`, `check`, `error` (no legend)
- Gemini review image tags (`![medium](url)`) auto-cleaned to `[medium]` in dense mode
- Consecutive `\n\n` normalized to `\n` in dense mode
- Errors go to stderr in `ERR\tkind\tmessage` format
- No ANSI colors

## Exit Codes

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
| `--ref REF` | Target ref for `file get` | Repository default branch |
| `--limit N` | Max results | 30 |
| `--max-files N` | Max file rows in `pr summary` | 15 |
| `--max-comments N` | Max comment rows in `pr summary` | 10 |
| `--state open\|closed\|all` | Filter state | open |
| `--checks` | Include CI check results in `pr summary` | off |
| `--full` | Verbose output (no abbreviations) | off |

## Design

- REST API v3 only (no GraphQL dependency)
- Standard library only (`net/http` + `encoding/json`)
- Static binary, zero external dependencies
- Token efficiency: 60-90% fewer tokens than `gh` JSON output
- Dense output minimizes both output tokens and LLM reasoning/thinking tokens
- `gh` CLI is optional (token fallback only)
- URL dispatch: `llmgh url <github-url>` auto-routes to the right command

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

## Related Projects

- [llmls](https://github.com/Kimi-Yori/llmls) — LLM-optimized file listing CLI
- [nanojq](https://github.com/Kimi-Yori/nanojq) — ultra-lightweight JSON selector for LLM pipelines
- [smart-trunc](https://github.com/Kimi-Yori/smart-trunc) — output truncation for LLM agents
- [cachit](https://github.com/Kimi-Yori/cachit) — command result cache for LLM agents
- [necache](https://github.com/Kimi-Yori/necache) — negative knowledge cache for LLM search

## License

MIT
