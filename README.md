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

func example() {
	println("covered")
	if false {
		println("not covered")
	}
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
package example

<C>func example() {
<C>     println("covered")
<N>     if false {
<N>             println("not covered")
        }
}
```

## Help

```
Usage of go-cover-view:
  -covered string
        prefix for covered line (default "<C>")
  -report string
        coverage report path (default "coverage.txt")
  -uncovered string
        prefix for uncovered line (default "<N>")
```

## License

MIT

## Author

Mitsuo Heijo (@johejo)
