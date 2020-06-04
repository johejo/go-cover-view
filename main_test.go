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

func Test_main(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("tmp=%s", tmp)
	t.Cleanup(func() {
		if err := os.RemoveAll(tmp); err != nil {
			t.Fatal(err)
		}
	})
	if err := cmdRun(tmp, "go", "mod", "init", "example.com/example"); err != nil {
		t.Fatal(err)
	}

	example := []byte(strings.TrimPrefix(`
package example

func example() {
	println("covered")
	if false {
		println("not covered")
	}
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

	if err := cmdRun(tmp, "go", "test", ".", "-v", "-cover", "-coverprofile", "coverage.txt"); err != nil {
		t.Fatal(err)
	}

	t.Run("render", func(t *testing.T) {
		if err := os.Chdir(tmp); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		w = &buf
		t.Cleanup(func() {
			w = os.Stdout
		})
		if err := _main(); err != nil {
			t.Fatal(err)
		}
		want := strings.TrimPrefix(`
example.com/example/example.go
  1: package example
  2: 
O 3: func example() {
O 4: 	println("covered")
X 5: 	if false {
X 6: 		println("not covered")
  7: 	}
  8: }

`, "\n")
		assert.Equal(t, want, buf.String())
	})

	t.Run("json", func(t *testing.T) {
		if err := os.Chdir(tmp); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		w = &buf
		t.Cleanup(func() {
			w = os.Stdout
		})
		_json = true
		if err := _main(); err != nil {
			t.Fatal(err)
		}
		assert.JSONEq(t, `{"example.com/example/example.go": {"coveredLines": [3, 4], "uncoveredLines": [5, 6]}}`, buf.String())
	})
}

func cmdRun(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = dir
	return cmd.Run()
}
