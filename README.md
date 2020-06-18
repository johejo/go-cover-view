# go-cover-view


[![ci](https://github.com/johejo/go-cover-view/workflows/ci/badge.svg?branch=master)](https://github.com/johejo/go-cover-view/actions?query=workflow%3Aci)
[![codecov](https://codecov.io/gh/johejo/go-cover-view/branch/master/graph/badge.svg)](https://codecov.io/gh/johejo/go-cover-view)
[![Go Report Card](https://goreportcard.com/badge/github.com/johejo/go-cover-view)](https://goreportcard.com/report/github.com/johejo/go-cover-view)

simple go coverage report viewer

## Install

```
go get github.com/johejo/go-cover-view
go install github.com/johejo/go-cover-view
```

## Get Started

Create a new go module.
```sh
mkdir -p example
cd example/
go mod init example.com/example
```

Create a Go file.

example.go
```go
package example

func example() bool {
	println("covered")
	if false {
		println("not covered")
	}
	return true
}
```

Create a Go test file.

example_test.go
```go
package example

import "testing"

func Test_example(t *testing.T) {
	example()
}
```

Run test and generate coverage report.

```sh
go test . -cover -coverprofile coverage.txt
```

view coverage report.

```sh
go-cover-view
```

```
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
```

json output

```sh
go-cover-view -output json
```

```json
[
  {
    "fileName": "example.com/example/example.go",
    "coveredLines": [3, 4, 8], 
    "uncoveredLines": [5, 6, 7]
  }
]
```

markdown output

```sh
go-cover-view -output markdown
```

```markdown
# Coverage Report


<details> <summary> example.com/example/example.go </summary>

\```
  1: package example
  2:
O 3: func example() bool {
O 4: 	println("covered")
X 5: 	if false {
X 6: 		println("not covered")
X 7: 	}
O 8: 	return true
  9: }
\```

</details>
```

## GitHub Actions Integration

Pull Request comment with markdown report

![Screenshot](https://user-images.githubusercontent.com/25817501/85069957-f1c8f300-b1ef-11ea-9ea3-7200e26483da.png)

GitHub Actions workflow example

```yaml
name: ci
on:
  pull_request:
    branches:
      - master
jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest]
        go: ["1.14"]
    runs-on: ${{ matrix.os }}
    timeout-minutes: 10
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: ${{ matrix.go }}

      - name: "test"
        run: |
          go test -cover -coverprofile coverage.txt -race -v ./...

      - name: "pull request comment"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          git fetch origin master
          go get github.com/johejo/go-cover-view
          go install github.com/johejo/go-cover-view
          go-cover-view -ci github-actions -git-diff-base origin/master
```

## Help

```
Usage of go-cover-view:
  -ci string
        ci type: available values "", "github-actions"
        github-actions:
                Comment the markdown report to Pull Request on GitHub.
  -covered string
        prefix for covered line (default "O")
  -git-diff-base string
        git diff base (default "origin/master")
  -git-diff-only
        only files with git diff
  -modfile string
        go.mod path
  -output string
        output type: available values "simple", "json", "markdown" (default "simple")
  -report string
        coverage report path (default "coverage.txt")
  -uncovered string
        prefix for uncovered line (default "X")
```

## License

MIT

## Author

Mitsuo Heijo (@johejo)
