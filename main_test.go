package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) string {
	t.Helper()
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("tmp=%s", tmp)
	t.Cleanup(func() {
		os.RemoveAll(tmp)
	})
	if err := cmdRun(tmp, "go", "mod", "init", "example.com/example"); err != nil {
		t.Fatal(err)
	}

	example := []byte(strings.TrimSpace(`
package example

func example() bool {
	println("covered")
	if false {
		println("not covered")
	}
	return true
}
`))
	if err := ioutil.WriteFile(filepath.Join(tmp, "example.go"), example, 0644); err != nil {
		t.Fatal(err)
	}

	exampleTest := []byte(strings.TrimSpace(`
package example

import "testing"

func Test_example(t *testing.T) {
	example()
}
`))
	if err := ioutil.WriteFile(filepath.Join(tmp, "example_test.go"), exampleTest, 0644); err != nil {
		t.Fatal(err)
	}

	return tmp
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	})
}

func Test_main(t *testing.T) {
	tmp := setup(t)
	if err := cmdRun(tmp, "go", "test", ".", "-v", "-cover", "-coverprofile", "coverage.txt"); err != nil {
		t.Fatal(err)
	}
	if err := cmdRun(tmp, "git", "init"); err != nil {
		t.Fatal(err)
	}

	t.Run("simple", func(t *testing.T) {
		want := readFileString(t, "testdata/output.txt")
		chdir(t, tmp)
		var buf bytes.Buffer
		writer = &buf
		t.Cleanup(func() {
			writer = os.Stdout
		})
		main()
		assert.Equal(t, want, strings.TrimSpace(buf.String()))
	})

	t.Run("json", func(t *testing.T) {
		want := readFileString(t, "testdata/output.json")
		chdir(t, tmp)
		var buf bytes.Buffer
		writer = &buf
		t.Cleanup(func() {
			writer = os.Stdout
		})
		output = "json"
		t.Cleanup(func() { output = "" })
		main()
		assert.JSONEq(t, string(want), buf.String())
	})

	t.Run("markdown", func(t *testing.T) {
		want := readFileString(t, "testdata/output.md")
		chdir(t, tmp)
		var buf bytes.Buffer
		writer = &buf
		t.Cleanup(func() {
			writer = os.Stdout
		})
		output = "markdown"
		t.Cleanup(func() { output = "" })
		main()
		assert.Equal(t, want, strings.TrimSpace(buf.String()))
	})

	t.Run("github-actions pull request comment", func(t *testing.T) {
		if testing.Short() {
			t.Skip()
		}
		if ci, _ := strconv.ParseBool(os.Getenv("CI")); !ci {
			t.Skip()
		}
		if commentTest, _ := strconv.ParseBool(os.Getenv("PULL_REQUEST_COMMENT_TEST")); !commentTest {
			t.Skip()
		}
		ci = "github-actions"
		gitDiffBase = ""
		t.Cleanup(func() { ci = ""; gitDiffBase = "origin/master" })
		main()
	})
}

func readFileString(t *testing.T, filename string) string {
	t.Helper()
	_want, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(strings.TrimSpace(string(_want)), "\r\n", "\n")
}

func cmdRun(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = dir
	return cmd.Run()
}

func Test_parseModfile(t *testing.T) {
	tmp := setup(t)
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	t.Run("empty", func(t *testing.T) {
		m, err := parseModfile()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, m.Path(), "example.com/example")
	})

	t.Run("specified", func(t *testing.T) {
		modFile = "./go.mod"
		t.Cleanup(func() { modFile = "" })
		m, err := parseModfile()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, m.Path(), "example.com/example")
	})
}

func Test_containsDiff(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		path     string
		diffs    []string
		want     bool
	}{
		{
			name:     "contains",
			filename: "example.com/example/main.go",
			path:     "example.com/example",
			diffs:    []string{"a.go", "main.go"},
			want:     true,
		},
		{
			name:     "not contains",
			filename: "example.com/example/main.go",
			path:     "example.com/example",
			diffs:    []string{"a.go", "b.go"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsDiff(tt.filename, tt.path, tt.diffs)
			assert.Equal(t, tt.want, got)
		})
	}
}
