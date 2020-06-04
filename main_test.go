package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	var buf bytes.Buffer
	out = &buf
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	if err := cmdRun(tmp, "go", "mod", "init", "example.com/example"); err != nil {
		t.Fatal(err)
	}

	example := []byte(`
package example

func example() {
	println("covered")
	if false {
		println("not covered")
	}
}
`)
	if err := ioutil.WriteFile(filepath.Join(tmp, "example.go"), example, 0644); err != nil {
		t.Fatal(err)
	}

	exampleTest := []byte(`
package example

import "testing"

func Test_example(t *testing.T) {
	example()
}
`)
	if err := ioutil.WriteFile(filepath.Join(tmp, "example_test.go"), exampleTest, 0644); err != nil {
		t.Fatal(err)
	}

	if err := cmdRun(tmp, "go", "test", ".", "-v", "-cover", "-coverprofile", "coverage.txt"); err != nil {
		t.Fatal(err)
	}

	if err := _main(); err != nil {
		t.Fatal(err)
	}

	const want = `
package example

<C>func example() {
<C>	println("covered")
<N>	if false {
<N>		println("not covered")
	}
}
`
	got := buf.String()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Log("want")
		t.Log(want)
		t.Log("got")
		t.Log(got)
		t.Log(diff)
		t.Fatal("unexpected output")
	}
}

func cmdRun(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = dir
	return cmd.Run()
}
