package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"golang.org/x/tools/cover"
)

var reportTmp = template.Must(template.New("report").Parse(`
# go-cover-view

{{range .}}
<details> <summary> {{.FileName}} </summary>
{{.Report}}
</details>
{{end}}
`))

type markdownRenderer struct{}

var _ renderer = (*markdownRenderer)(nil)

func (r *markdownRenderer) Render(w io.Writer, profiles []*cover.Profile, path string) error {
	results, err := getMarkdownReports(profiles, path)
	if err != nil {
		return err
	}
	return reportTmp.ExecuteTemplate(w, "report", results)
}

type markdownReport struct {
	FileName string
	Report   string
}

func newMarkdownReport(fileName string, lines []string) *markdownReport {
	return &markdownReport{
		FileName: fileName,
		Report:   buildReport(lines),
	}
}

func getMarkdownReports(profiles []*cover.Profile, path string) ([]*markdownReport, error) {
	diffs, err := getDiffs()
	if err != nil {
		return nil, err
	}
	reports := make([]*markdownReport, 0, len(profiles))
	for _, profile := range profiles {
		lines, err := getLines(profile, path)
		if err != nil {
			return nil, err
		}
		if gitDiffOnly {
			if containsDiff(profile.FileName, path, diffs) {
				reports = append(reports, newMarkdownReport(profile.FileName, lines))
			}
			continue
		}
		reports = append(reports, newMarkdownReport(profile.FileName, lines))
	}
	return reports, nil
}

func buildReport(lines []string) string {
	var b strings.Builder
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "```")
	for _, line := range lines {
		fmt.Fprintln(&b, line)
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "```")
	return b.String()
}

func upsertGitHubPullRequestComment(profiles []*cover.Profile, path string) error {
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return errors.New("env: GITHUB_EVENT_PATH is missing")
	}
	f, err := os.Open(eventPath)
	if err != nil {
		return err
	}
	defer f.Close()
	var event github.PullRequestEvent
	if err := json.NewDecoder(f).Decode(&event); err != nil {
		return err
	}
	ctx := context.Background()
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return errors.New("env: GITHUB_TOKEN is missing")
	}

	var buf bytes.Buffer
	r := &markdownRenderer{}
	if err := r.Render(&buf, profiles, path); err != nil {
		return err
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, ts)
	gc := github.NewClient(httpClient)
	pr := event.GetPullRequest()
	_repo := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	if len(_repo) != 2 {
		return fmt.Errorf("invalid env: GITHUB_REPOSITORY=%v", _repo)
	}
	owner := _repo[0]
	repo := _repo[1]
	comments, _, err := gc.Issues.ListComments(ctx, owner, repo, pr.GetNumber(), nil)
	if err != nil {
		return err
	}
	var commentID int64
	for _, c := range comments {
		u := c.GetUser()
		if u.GetLogin() == "github-actions[bot]" && u.GetType() == "Bot" && strings.Contains(c.GetBody(), "# go-cover-view") {
			commentID = c.GetID()
			break
		}
	}
	body := buf.String()
	if commentID == 0 {
		_, _, err := gc.Issues.CreateComment(ctx, owner, repo, pr.GetNumber(), &github.IssueComment{
			Body: &body,
		})
		if err != nil {
			return err
		}
	} else {
		_, _, err := gc.Issues.EditComment(ctx, owner, repo, commentID, &github.IssueComment{
			Body: &body,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
