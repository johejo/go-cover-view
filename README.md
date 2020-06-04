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
example.com/example/example.go
  1: package example
  2: 
O 3: func example() {
O 4: 	println("covered")
X 5: 	if false {
X 6: 		println("not covered")
  7: 	}
  8: }

```

## Help

```
Usage of go-cover-view:
  -covered string
        prefix for covered line (default "O")
  -json
        json output
  -report string
        coverage report path (default "coverage.txt")
  -uncovered string
        prefix for uncovered line (default "X")
```

## License

MIT

## Author

Mitsuo Heijo (@johejo)
