package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/tools/cover"
)

var (
	report    string
	covered   string
	uncovered string
	_json     bool

	w io.Writer = os.Stdout
)

type modfile struct {
	Module struct {
		Path string
	}
}

func init() {
	flag.StringVar(&report, "report", "coverage.txt", "coverage report path")
	flag.StringVar(&covered, "covered", "O", "prefix for covered line")
	flag.StringVar(&uncovered, "uncovered", "X", "prefix for uncovered line")
	flag.BoolVar(&_json, "json", false, "json output")
}

func main() {
	flag.Parse()
	if err := _main(); err != nil {
		log.Fatal(err)
	}
}

type result struct {
	CoveredLines   []int `json:"coveredLines"`
	UncoveredLines []int `json:"uncoveredLines"`
}

func _main() error {
	output, err := exec.Command("go", "mod", "edit", "-json").Output()
	if err != nil {
		return err
	}

	var m modfile
	if err := json.Unmarshal(output, &m); err != nil {
		return err
	}

	profiles, err := cover.ParseProfiles(report)
	if err != nil {
		return err
	}
	results := make(map[string]result, len(profiles))
	for _, profile := range profiles {
		lines, err := getLines(profile, m.Module.Path)
		if err != nil {
			return err
		}
		if _json {
			coveredLines := make([]int, 0, len(lines))
			uncoveredLines := make([]int, 0, len(lines))
			for i, line := range lines {
				switch {
				case strings.HasPrefix(line, covered):
					coveredLines = append(coveredLines, i+1)
				case strings.HasPrefix(line, uncovered):
					uncoveredLines = append(uncoveredLines, i+1)
				}
			}
			results[profile.FileName] = result{
				CoveredLines:   coveredLines,
				UncoveredLines: uncoveredLines,
			}
			continue
		}
		fmt.Fprintln(w, profile.FileName)
		fmt.Fprintln(w, strings.Join(lines, "\n"))
		fmt.Fprintln(w)
	}
	if _json {
		return json.NewEncoder(w).Encode(results)
	}
	return nil
}

func getLines(profile *cover.Profile, module string) ([]string, error) {
	p := strings.TrimPrefix(profile.FileName, module+"/")
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
		format := "%" + w + "d: %s"
		newLine := fmt.Sprintf(format, i+1, line)
		lines[i] = newLine
	}

	for _, block := range profile.Blocks {
		var prefix string
		if block.Count > 0 {
			prefix = covered
		} else {
			prefix = uncovered
		}
		for i := block.StartLine - 1; i < block.EndLine-1; i++ {
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
