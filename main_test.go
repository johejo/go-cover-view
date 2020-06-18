package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

	example := []byte(strings.TrimPrefix(`
package example

func example() bool {
	println("covered")
	if false {
		println("not covered")
	}
	return true
}
`, "\n"))
	if err := ioutil.WriteFile(filepath.Join(tmp, "example.go"), example, 0644); err != nil {
		t.Fatal(err)
	}

	exampleTest := []byte(strings.TrimPrefix(`
package example

import "testing"

func Test_example(t *testing.T) {
	example()
}
`, "\n"))
	if err := ioutil.WriteFile(filepath.Join(tmp, "example_test.go"), exampleTest, 0644); err != nil {
		t.Fatal(err)
	}

	return tmp
}

func Test_main(t *testing.T) {
	tmp := setup(t)
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	if err := cmdRun(tmp, "go", "test", ".", "-v", "-cover", "-coverprofile", "coverage.txt"); err != nil {
		t.Fatal(err)
	}

	t.Run("render", func(t *testing.T) {
		var buf bytes.Buffer
		w = &buf
		t.Cleanup(func() {
			w = os.Stdout
		})
		main()
		want := strings.TrimPrefix(`
example.com/example/example.go
  1: package example
  2: 
O 3: func example() bool {
O 4: 	println("covered")
X 5: 	if false {
X 6: 		println("not covered")
X 7: 	}
O 8: 	return true
  9: }

`, "\n")
		assert.Equal(t, want, buf.String())
	})

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		w = &buf
		t.Cleanup(func() {
			w = os.Stdout
		})
		_json = true
		t.Cleanup(func() { _json = false })
		main()
		// language=json
		const want = `
[
  {
    "fileName": "example.com/example/example.go",
    "coveredLines": [3, 4, 8], 
    "uncoveredLines": [5, 6, 7]
  }
]
`
		assert.JSONEq(t, want, buf.String())
	})
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
