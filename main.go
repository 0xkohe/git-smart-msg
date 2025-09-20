package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	openai "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
        "github.com/openai/openai-go/v2/shared"
)

// ============================
// Types
// ============================

type PlanItem struct {
	SHA         string `json:"sha"`
	OldMessage  string `json:"old_message"`
	NewMessage  string `json:"new_message"`
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
	AuthorDate  string `json:"author_date"` // RFC3339
}

type Plan struct {
	RepoPath    string     `json:"repo_path"`
	Base        string     `json:"base"` // exclusive (parent side), empty means computed
	Head        string     `json:"head"` // inclusive tip
	CreatedAt   string     `json:"created_at"`
	Model       string     `json:"model"`
	AllowMerges bool       `json:"allow_merges"`
	Items       []PlanItem `json:"items"`
}

type AIClient interface {
	SuggestMessage(ctx context.Context, model string, diff string, oldMsg string) (string, error)
}

// ============================
// OpenAI SDK Client (v2)
// ============================

type OpenAIClient struct {
	client openai.Client
}

func NewOpenAIClient() (*OpenAIClient, error) {
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY is not set")
	}
	base := strings.TrimSpace(os.Getenv("OPENAI_API_BASE"))

	var opts []option.RequestOption
	opts = append(opts, option.WithAPIKey(apiKey))
	if base != "" {
		opts = append(opts, option.WithBaseURL(base))
	}

	cli := openai.NewClient(opts...)
	return &OpenAIClient{client: cli}, nil
}

func (c *OpenAIClient) SuggestMessage(ctx context.Context, model string, diff string, oldMsg string) (string, error) {
	sys := `You are an expert at writing precise, helpful Git commit messages.
Follow the "Conventional Commits" style when appropriate.
One short summary line (<= 72 chars), then an empty line, then bullet points if needed.
Use imperative present tense (e.g., "fix: handle nil pointer in X").
If the diff is large, summarize purpose + major changes concisely.`

	user := fmt.Sprintf(
		"Old message:\n\"%s\"\n\nDiff (unified, files & hunks):\n%s",
		oldMsg, truncate(diff, 40000),
	)

	params := openai.ChatCompletionNewParams{
		Model: shared.ChatModel(model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(sys),
			openai.UserMessage(user),
		},
		MaxCompletionTokens:  openai.Int(4000),
	}

	resp, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("no choices returned")
	}

	// v2 SDKは Content を stringで保持（README参照）
	txt := strings.TrimSpace(resp.Choices[0].Message.Content)
	txt = strings.Trim(txt, "` \n")
	if txt == "" {
		return "", errors.New("empty content")
	}
	return txt, nil
}

// ============================
// Git helpers
// ============================

func git(args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %v failed: %v, %s", args, err, stderr.String())
	}
	return stdout.String(), nil
}

func ensureCleanWorktree() error {
	out, err := git("status", "--porcelain")
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) != "" {
		return errors.New("worktree is not clean; commit/stash first")
	}
	return nil
}

type CommitMeta struct {
	SHA         string
	Subject     string
	AuthorName  string
	AuthorEmail string
	AuthorDate  time.Time
	IsMerge     bool
}

func listCommits(rangeExpr string) ([]CommitMeta, error) {
	// %H SHA, %s subject, %an, %ae, %ad (ISO8601), %P parents
	format := "%H%x1f%s%x1f%an%x1f%ae%x1f%aI%x1f%P%x1e"
	out, err := git("log", "--reverse", "--format="+format, rangeExpr)
	if err != nil {
		return nil, err
	}
	var commits []CommitMeta
	records := strings.Split(strings.TrimSuffix(out, "\x1e"), "\x1e")
	for _, rec := range records {
		if strings.TrimSpace(rec) == "" {
			continue
		}
		parts := strings.Split(rec, "\x1f")
		if len(parts) < 6 {
			continue
		}
		dt, _ := time.Parse(time.RFC3339, parts[4])

		parents := strings.Fields(parts[5])
		isMerge := len(parents) > 1

		commits = append(commits, CommitMeta{
			SHA:         strings.TrimSpace(parts[0]),
			Subject:     parts[1],
			AuthorName:  parts[2],
			AuthorEmail: parts[3],
			AuthorDate:  dt,
			IsMerge:     isMerge,
		})
	}
	return commits, nil
}

func showDiff(sha string) (string, error) {
	// ユニファイド差分（空白無視はしない/正確さ優先）
	out, err := git("show", "--patch", "--unified=3", "--no-color", "--find-renames", sha)
	if err != nil {
		return "", err
	}
	return out, nil
}

// ============================
// Utilities
// ============================

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "\n...[truncated]..."
}

func repoTop() (string, error) {
	out, err := git("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func defaultHead() (string, error) {
	out, err := git("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func nthAncestor(head string, n int) (string, error) {
	spec := fmt.Sprintf("%s~%d", head, n)
	out, err := git("rev-parse", spec)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// ============================
// Plan command
// ============================

func cmdPlan(args []string) error {
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	limit := fs.Int("limit", 20, "number of commits from HEAD to include")
	rangeExpr := fs.String("range", "", "explicit git range (e.g., <base>..<head>)")
	model := fs.String("model", envOr("OPENAI_MODEL", "gpt-5-nano"), "LLM model")
	allowMerges := fs.Bool("allow-merges", false, "include merge commits (not recommended)")
	outFile := fs.String("out", "plan.json", "output plan file")
	timeout := fs.Duration("timeout", 25*time.Second, "per-commit AI timeout")
	fs.Parse(args)

	head, err := defaultHead()
	if err != nil {
		return err
	}
	base := ""
	if *rangeExpr == "" {
		anc, err := nthAncestor(head, *limit)
		if err != nil {
			ancOut, err2 := git("rev-list", "--max-parents=0", "HEAD")
			if err2 != nil {
				return fmt.Errorf("cannot compute base: %v, %v", err, err2)
			}
			anc = strings.TrimSpace(ancOut)
		}
		base = anc
		*rangeExpr = fmt.Sprintf("%s..%s", base, head)
	}

	commits, err := listCommits(*rangeExpr)
	if err != nil {
		return err
	}
	if len(commits) == 0 {
		return errors.New("no commits in range")
	}

	ai, err := NewOpenAIClient()
	if err != nil {
		return err
	}

	var items []PlanItem
	for _, c := range commits {
		if c.IsMerge && !*allowMerges {
			log.Printf("skip merge commit %s", c.SHA)
			continue
		}
		diff, err := showDiff(c.SHA)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		newMsg, err := ai.SuggestMessage(ctx, *model, diff, c.Subject)
		cancel()
		if err != nil {
			return fmt.Errorf("AI failed for %s: %w", c.SHA, err)
		}
		items = append(items, PlanItem{
			SHA:         c.SHA,
			OldMessage:  c.Subject,
			NewMessage:  sanitizeMessage(newMsg),
			AuthorName:  c.AuthorName,
			AuthorEmail: c.AuthorEmail,
			AuthorDate:  c.AuthorDate.Format(time.RFC3339),
		})
		log.Printf("planned: %s  %s  ->  %s", c.SHA[:7], truncate(c.Subject, 60), truncate(newMsg, 60))
	}

	top, _ := repoTop()
	plan := Plan{
		RepoPath:    top,
		Base:        base,
		Head:        head,
		CreatedAt:   time.Now().Format(time.RFC3339),
		Model:       *model,
		AllowMerges: *allowMerges,
		Items:       items,
	}
	data, _ := json.MarshalIndent(plan, "", "  ")
	if err := os.WriteFile(*outFile, data, 0644); err != nil {
		return err
	}
	fmt.Printf("Wrote %s (%d messages)\n", *outFile, len(items))
	return nil
}

func sanitizeMessage(s string) string {
	// 先頭行の長さを72字程度に抑える（切り捨てはしない、整形のみ）
	lines := splitLines(s)
	if len(lines) == 0 {
		return "chore: update"
	}
	first := strings.TrimSpace(lines[0])
	first = regexp.MustCompile(`^\[(feat|fix|docs|style|refactor|perf|test|chore)\]\s*:`).ReplaceAllString(first, "$1:")
	rest := strings.Join(lines[1:], "\n")
	first = strings.Trim(first, "# ")
	msg := first
	if strings.TrimSpace(rest) != "" {
		msg += "\n\n" + strings.TrimSpace(rest)
	}
	return msg
}

func splitLines(s string) []string {
	return regexp.MustCompile(`\r?\n`).Split(s, -1)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// ============================
// Apply command (linear history only)
// ============================

func cmdApply(args []string) error {
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	inFile := fs.String("in", "plan.json", "plan file path")
	newBranch := fs.String("branch", "", "new branch to create (required)")
	allowMerges := fs.Bool("allow-merges", false, "attempt to preserve merge commits (best-effort; otherwise abort)")
	fs.Parse(args)

	if *newBranch == "" {
		return errors.New("--branch is required")
	}

	if err := ensureCleanWorktree(); err != nil {
		return err
	}
	var plan Plan
	b, err := os.ReadFile(*inFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &plan); err != nil {
		return err
	}
	if len(plan.Items) == 0 {
		return errors.New("plan has no items")
	}

	// 作業ブランチ
	if _, err := git("checkout", "-b", *newBranch); err != nil {
		return err
	}
	// 起点を base にリセット
	base := plan.Base
	if strings.TrimSpace(base) == "" {
		first := plan.Items[0].SHA
		parent, err := git("rev-parse", first+"^")
		if err != nil {
			return fmt.Errorf("cannot determine base: %w", err)
		}
		base = strings.TrimSpace(parent)
	}
	if _, err := git("reset", "--hard", base); err != nil {
		return err
	}

	// cherry-pick で1件ずつ適用
	for _, it := range plan.Items {
		if !*allowMerges {
			parents, _ := git("rev-list", "--parents", "-n", "1", it.SHA)
			if strings.Count(strings.TrimSpace(parents), " ") >= 2 {
				return fmt.Errorf("merge commit detected (%s). rerun with --allow-merges (experimental).", it.SHA[:7])
			}
		}

		if _, err := git("cherry-pick", "-n", it.SHA); err != nil {
			_, _ = git("cherry-pick", "--abort")
			return fmt.Errorf("cherry-pick failed at %s; resolve manually and rerun", it.SHA[:7])
		}

		authorFlag := fmt.Sprintf("--author=%s <%s>", it.AuthorName, it.AuthorEmail)
		commitEnv := os.Environ()
		commitEnv = append(commitEnv,
			"GIT_COMMITTER_NAME="+it.AuthorName,
			"GIT_COMMITTER_EMAIL="+it.AuthorEmail,
			"GIT_COMMITTER_DATE="+it.AuthorDate,
			"GIT_AUTHOR_DATE="+it.AuthorDate,
		)

		msg := it.NewMessage
		if strings.TrimSpace(msg) == "" {
			msg = it.OldMessage
		}

		diffIndex, _ := git("diff", "--cached", "--name-only")
		if strings.TrimSpace(diffIndex) == "" {
			log.Printf("skip empty commit %s", it.SHA[:7])
			_, _ = git("reset")
			continue
		}

		var stdout, stderr bytes.Buffer
		cmd := exec.Command("git", "commit", "-m", msg, authorFlag, "--no-verify")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmd.Env = commitEnv
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git commit failed: %v, %s", err, stderr.String())
		}
		log.Printf("rewritten: %s", it.SHA[:7])
	}

	fmt.Printf("\n✅ Done. New branch %q contains rewritten history.\n", *newBranch)
	fmt.Println("⚠️  Rewriting history rewrites SHAs. Coordinate with your team before force-pushing:")
	fmt.Printf("   git push --force-with-lease origin %s\n", *newBranch)
	return nil
}

// ============================
// main
// ============================

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, `git-smartmsg

Subcommands:
  plan   - generate AI commit messages for a range (writes plan.json)
  apply  - apply plan.json on a new branch as rewritten linear history

Examples:
  git-smartmsg plan --limit 30 --model gpt-5-nano
  git-smartmsg apply --branch rewrite/2025-09-20
`)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "plan":
		if err := cmdPlan(os.Args[2:]); err != nil {
			log.Fatal("plan error: ", err)
		}
	case "apply":
		if err := cmdApply(os.Args[2:]); err != nil {
			log.Fatal("apply error: ", err)
		}
	default:
		log.Fatal("unknown subcommand")
	}
}

