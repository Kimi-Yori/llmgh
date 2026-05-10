<table>
    <thead>
        <tr>
            <th style="text-align:center"><a href="README.md">English</a></th>
            <th style="text-align:center">日本語</th>
        </tr>
    </thead>
</table>

# llmgh

LLMエージェント向けのコンパクトなGitHubコンテキストリーダー。

llmghは`gh`の代替ではありません。GitHubのデータ（PR、Issue、ファイル、チェック）を読み取り、トークン効率に最適化された安定したTSV形式で出力します。Claude Code、Cursor、Codex、その他のLLMコーディングエージェント向けに設計されています。

## なぜllmgh？

`gh pr view`は25行以上の冗長な人間向けテキストを出力します。LLMエージェントは出力トークンだけでなく、それを解析するための推論トークンにもコストがかかります。

llmghは情報の欠落なしに60-90%のトークン削減を実現します。

## インストール

**ビルド済みバイナリ:**

```bash
# Linux x86_64
curl -L -o llmgh https://github.com/Kimi-Yori/llmgh/releases/latest/download/llmgh-linux-amd64

# macOS Apple Silicon (M1/M2/M3/M4)
curl -L -o llmgh https://github.com/Kimi-Yori/llmgh/releases/latest/download/llmgh-darwin-arm64

# macOS Intel
curl -L -o llmgh https://github.com/Kimi-Yori/llmgh/releases/latest/download/llmgh-darwin-amd64

chmod +x llmgh
sudo mv llmgh /usr/local/bin/
```

**ソースからビルド**（Go 1.22+が必要）:

```bash
git clone https://github.com/Kimi-Yori/llmgh.git
cd llmgh
make build
sudo make install
```

**クロスビルド:**

```bash
make release    # linux/amd64, darwin/amd64, darwin/arm64
# 出力: dist/llmgh-{os}-{arch}
```

## 認証

トークン解決順序:

1. `LLMGH_TOKEN`
2. `GH_TOKEN`
3. `GITHUB_TOKEN`
4. `gh auth token`（フォールバック、gh CLIが必要）

## コマンド

### url

```bash
llmgh url https://github.com/owner/repo/pull/123#pullrequestreview-456
```

GitHub URLを解釈して、対応するサブコマンド（`status` / `pr` / `issue` / `file`）に振り分けます。

`blob`/`tree` URLは`file get`に振り分けます。`blob`/`tree`直後の1セグメントをrefとして扱います。`feature/foo`のような`/`を含むrefでは、明示的に`file get`を使ってください:

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
llmgh pr summary 123 --repo owner/repo --checks          # CI結果も含める
llmgh pr summary 123 --repo owner/repo --max-files 5 --max-comments 3
```

Dense出力（デフォルト）:
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

GitHub Contents APIのrawモードでファイル内容を標準出力します。パスの各セグメントはURLエンコードされるため、スペースや`?`/`#`を含むファイル名も扱えます。

## 出力フォーマット

- TSV（タブ区切り）、デフォルトはdense（`--full`で冗長版）
- 1レコード1行
- フィールド値内のタブ/改行/復帰は`\t`、`\n`、`\r`にエスケープ
- summaryコマンドは先頭に`legend`行を出力（denseのみ）
- Dense略称: `merg`(mergeable), `cmt`(comment), `rc`(review_comment), `chk`(check), `err`(error)
- Full出力（`--full`）: `mergeable`, `comment`, `review_comment`, `check`, `error`（legendなし）
- Geminiレビュー画像タグ（`![medium](url)`）はdenseモードで`[medium]`に自動クリーニング
- 連続`\n\n`はdenseモードで`\n`に正規化
- エラーはstderrに`ERR\tkind\tmessage`形式で出力
- ANSIカラーなし

## 終了コード

| コード | 意味 |
|--------|------|
| 0 | 成功 |
| 1 | 使用法エラー |
| 2 | 認証エラー |
| 3 | 見つからない |
| 4 | レート制限 |
| 5 | ネットワークエラー |
| 6 | APIエラー |

## オプション

| フラグ | 説明 | デフォルト |
|--------|------|-----------|
| `--repo owner/repo` | 対象リポジトリ | git remoteから検出 |
| `--ref REF` | `file get`の対象ref | リポジトリのデフォルトブランチ |
| `--limit N` | 最大結果数 | 30 |
| `--max-files N` | `pr summary`の最大ファイル行数 | 15 |
| `--max-comments N` | `pr summary`の最大コメント行数 | 10 |
| `--state open\|closed\|all` | 状態フィルタ | open |
| `--checks` | `pr summary`にCIチェック結果を含める | off |
| `--full` | 冗長出力（略称なし） | off |

## 設計

- REST API v3のみ（GraphQL依存なし）
- 標準ライブラリのみ（`net/http` + `encoding/json`）
- スタティックバイナリ、外部依存ゼロ
- トークン効率: `gh` JSON出力比60-90%削減
- Dense出力はLLMの出力トークンと推論/思考トークンの両方を最小化
- `gh` CLIはオプション（トークンのフォールバックのみ）
- URL振り分け: `llmgh url <github-url>`で適切なコマンドに自動ルーティング

## 比較

```
# gh pr view（GraphQL、冗長）
Title:  Fix auth flow
State:  OPEN
Author: alice
Labels: bug, security
Assignees: bob
...25行以上

# llmgh pr view（REST、コンパクト）
pr	123	open	main	feature/foo	alice	Fix auth flow	2026-04-23T10:12:00Z
mergeable	true	draft=false	comments=3	changed_files=5	+120	-30
labels	bug,security
...4行
```

## 関連プロジェクト

- [llmls](https://github.com/Kimi-Yori/llmls) — LLM最適化ファイルリスティングCLI
- [nanojq](https://github.com/Kimi-Yori/nanojq) — LLMパイプライン向け超軽量JSONセレクタ
- [smart-trunc](https://github.com/Kimi-Yori/smart-trunc) — LLMエージェント向け出力トランケーション
- [cachit](https://github.com/Kimi-Yori/cachit) — LLMエージェント向けコマンド結果キャッシュ
- [necache](https://github.com/Kimi-Yori/necache) — LLM検索向け否定知識キャッシュ

## ライセンス

MIT
