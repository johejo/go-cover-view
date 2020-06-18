package main

import (
	"io"
	"strings"
	"text/template"

	"golang.org/x/tools/cover"
)

var reportTmp = template.Must(template.New("report").Parse(`
# Coverage Report

{{range .}}
<details> <summary> {{.FileName}} </summary>
{{.Report}}
</details>
{{end}}
`))

type markdownRenderer struct {
}

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
	b.WriteString("\n```\n")
	for _, line := range lines {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n```\n")
	return b.String()
}
