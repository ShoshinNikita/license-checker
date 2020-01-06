# License Checker

Check licenses of dependencies

## Installation

```
go get github.com/ShoshinNikita/license-checker
```

## Process

1. Get list of all dependencies with `go list -m all` command
2. Check licenses with [go.dev](https://go.dev/) service. License can be found and parsed in **License** tab ([example](https://pkg.go.dev/github.com/pkg/errors?tab=licenses))
3. Print licenses and modules. For example:

```
List of licenses:
! unknown:
  - github.com/davecgh/go-spew@v1.1.0
Apache-2.0:
  - gopkg.in/yaml.v2@v2.2.4
BSD-2-Clause:
  - github.com/pkg/errors@v0.8.1
BSD-3-Clause:
  - github.com/golang/protobuf@v1.3.2
MIT:
  - github.com/caarlos0/env/v6@v6.1.0
  - github.com/fatih/color@v1.8.0
  - github.com/ShoshinNikita/go-clog/v3@v3.2.0
```
