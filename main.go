package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/tools/cover"
)

var (
	report    string
	covered   string
	uncovered string

	out io.Writer = os.Stdout
)

type modfile struct {
	Module struct {
		Path string
	}
}

func init() {
	flag.StringVar(&report, "report", "coverage.txt", "coverage report path")
	flag.StringVar(&covered, "covered", "<C>", "prefix for covered line")
	flag.StringVar(&uncovered, "uncovered", "<N>", "prefix for uncovered line")
}

func main() {
	flag.Parse()
	if err := _main(); err != nil {
		log.Fatal(err)
	}
}

func _main() error {
	out, err := exec.Command("go", "mod", "edit", "-json").Output()
	if err != nil {
		return err
	}

	var m modfile
	if err := json.Unmarshal(out, &m); err != nil {
		return err
	}

	profiles, err := cover.ParseProfiles(report)
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		if err := render(profile, m.Module.Path); err != nil {
			return err
		}
	}
	return nil
}

func render(profile *cover.Profile, module string) error {
	p := strings.TrimPrefix(profile.FileName, module+"/")
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lines := make([]string, 0, 1000)
	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()
		lines = append(lines, line)
	}

	for _, block := range profile.Blocks {
		var b strings.Builder
		if block.Count > 0 {
			b.WriteString(covered)
		} else {
			b.WriteString(uncovered)
		}
		prefix := b.String()
		for i := block.StartLine; i < block.EndLine; i++ {
			x := i - 1
			lines[x] = prefix + lines[x]
		}
	}

	fmt.Fprintln(out, strings.Join(lines, "\n"))
	return nil
}
