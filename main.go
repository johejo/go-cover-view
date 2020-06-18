package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/cover"
)

var (
	modFile   string
	report    string
	covered   string
	uncovered string

	output string
	ci     string

	gitDiffOnly bool
	gitDiffBase string

	writer io.Writer = os.Stdout
)

type _modfile interface {
	Path() string
}

type modfileFromJSON struct {
	Module struct {
		Path string
	}
}

func (m *modfileFromJSON) Path() string {
	return m.Module.Path
}

type xmodfile struct {
	*modfile.File
}

func (m *xmodfile) Path() string {
	return m.Module.Mod.Path
}

func init() {
	flag.StringVar(&modFile, "modfile", "", "go.mod path")
	flag.StringVar(&report, "report", "coverage.txt", "coverage report path")
	flag.StringVar(&covered, "covered", "O", "prefix for covered line")
	flag.StringVar(&uncovered, "uncovered", "X", "prefix for uncovered line")

	flag.StringVar(&output, "output", "simple", `output type: available values "simple", "json", "markdown"`)
	flag.StringVar(&ci, "ci", "", strings.TrimSpace(`
ci type: available values "", "github-actions"
github-actions:
	Comment the markdown report to Pull Request on GitHub.
`))

	flag.BoolVar(&gitDiffOnly, "git-diff-only", false, "only files with git diff")
	flag.StringVar(&gitDiffBase, "git-diff-base", "origin/master", "git diff base")
}

func main() {
	flag.Parse()
	if err := _main(); err != nil {
		log.Fatal(err)
	}
}

func _main() error {
	m, err := parseModfile()
	if err != nil {
		return err
	}

	profiles, err := cover.ParseProfiles(report)
	if err != nil {
		return err
	}

	switch ci {
	case "github-actions":
		gitDiffOnly = true
		return upsertGitHubPullRequestComment(profiles, m.Path())
	}

	var r renderer
	switch output {
	case "", "simple":
		r = &simpleRenderer{}
	case "json":
		r = &jsonRenderer{}
	case "markdown":
		r = &markdownRenderer{}
	default:
		return fmt.Errorf("invalid output type %s", output)
	}

	return r.Render(writer, profiles, m.Path())
}

func parseModfile() (_modfile, error) {
	if modFile == "" {
		output, err := exec.Command("go", "mod", "edit", "-json").Output()
		if err != nil {
			return nil, fmt.Errorf("go mod edit -json: %w", err)
		}
		var m modfileFromJSON
		if err := json.Unmarshal(output, &m); err != nil {
			return nil, err
		}
		return &m, nil
	}

	data, err := ioutil.ReadFile(modFile)
	if err != nil {
		return nil, err
	}

	f, err := modfile.Parse(modFile, data, nil)
	if err != nil {
		return nil, err
	}
	return &xmodfile{File: f}, nil
}

func getLines(profile *cover.Profile, module string) ([]string, error) {
	// github.com/johejo/go-cover-view/main.go -> ./main.go
	p := strings.ReplaceAll(profile.FileName, module, ".")
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lines := make([]string, 0, 1000)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	width := int(math.Log10(float64(len(lines)))) + 1
	if len(covered) > len(uncovered) {
		width += len(covered) + 1
	} else {
		width += len(uncovered) + 1
	}
	w := strconv.Itoa(width)
	for i, line := range lines {
		var newLine string
		if len(line) == 0 {
			format := "%" + w + "d:"
			newLine = fmt.Sprintf(format, i+1)
		} else {
			format := "%" + w + "d: %s"
			newLine = fmt.Sprintf(format, i+1, line)
		}
		lines[i] = newLine
	}

	for _, block := range profile.Blocks {
		var prefix string
		if block.Count > 0 {
			prefix = covered
		} else {
			prefix = uncovered
		}
		for i := block.StartLine - 1; i <= block.EndLine-1; i++ {
			if i >= len(lines) {
				return nil, fmt.Errorf("invalid line length: index=%d, len(lines)=%d", i, len(lines))
			}
			line := lines[i]
			newLine := prefix + line[len(prefix):]
			lines[i] = newLine
		}
	}

	return lines, nil
}

func getDiffs() ([]string, error) {
	if !gitDiffOnly {
		return []string{}, nil
	}
	args := []string{"diff", "--name-only"}
	if gitDiffBase != "" {
		args = append(args, gitDiffBase)
	}
	_out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, err
	}
	out := strings.TrimSpace(string(_out))
	diffs := strings.Split(out, "\n")
	return diffs, nil
}

func containsDiff(filename, path string, diffs []string) bool {
	for _, diff := range diffs {
		name := fmt.Sprintf("%s/%s", path, diff)
		if filename == name {
			return true
		}
	}
	return false
}

type renderer interface {
	Render(w io.Writer, profiles []*cover.Profile, path string) error
}

type simpleRenderer struct{}

var _ renderer = (*simpleRenderer)(nil)

func (r *simpleRenderer) Render(w io.Writer, profiles []*cover.Profile, path string) error {
	reports, err := getSimpleReports(profiles, path)
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(w)
	for _, r := range reports {
		fmt.Fprintln(bw, r.FileName)
		fmt.Fprintln(bw, r.Report)
		fmt.Fprintln(bw)
	}
	return bw.Flush()
}

type simpleReport struct {
	FileName string
	Report   string
}

func newSimpleReport(fileName string, lines []string) *simpleReport {
	return &simpleReport{
		FileName: fileName,
		Report:   strings.Join(lines, "\n"),
	}
}

func getSimpleReports(profiles []*cover.Profile, path string) ([]*simpleReport, error) {
	diffs, err := getDiffs()
	if err != nil {
		return nil, err
	}
	results := make([]*simpleReport, 0, len(profiles))
	for _, profile := range profiles {
		lines, err := getLines(profile, path)
		if err != nil {
			return nil, err
		}
		if gitDiffOnly {
			if containsDiff(profile.FileName, path, diffs) {
				results = append(results, newSimpleReport(profile.FileName, lines))
			}
			continue
		}
		results = append(results, newSimpleReport(profile.FileName, lines))
	}
	return results, nil
}
