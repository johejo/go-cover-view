package main

import (
	"encoding/json"
	"io"
	"strings"

	"golang.org/x/tools/cover"
)

type jsonRenderer struct{}

var _ renderer = (*jsonRenderer)(nil)

func (r *jsonRenderer) Render(w io.Writer, profiles []*cover.Profile, path string) error {
	results, err := getJSONResults(profiles, path)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(results)
}

type jsonResult struct {
	FileName       string `json:"fileName"`
	CoveredLines   []int  `json:"coveredLines"`
	UncoveredLines []int  `json:"uncoveredLines"`
}

func newJSONResult(fileName string, lines []string) *jsonResult {
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
	return &jsonResult{
		FileName:       fileName,
		CoveredLines:   coveredLines,
		UncoveredLines: uncoveredLines,
	}
}

func getJSONResults(profiles []*cover.Profile, path string) ([]*jsonResult, error) {
	diffs, err := getDiffs()
	if err != nil {
		return nil, err
	}
	results := make([]*jsonResult, 0, len(profiles))
	for _, profile := range profiles {
		lines, err := getLines(profile, path)
		if err != nil {
			return nil, err
		}
		if gitDiffOnly {
			if containsDiff(profile.FileName, path, diffs) {
				results = append(results, newJSONResult(profile.FileName, lines))
			}
		} else {
			results = append(results, newJSONResult(profile.FileName, lines))
		}
	}
	return results, nil
}
