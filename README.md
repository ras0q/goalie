# 🥅 Goalie

[![Go Reference](https://pkg.go.dev/badge/github.com/ras0q/goalie.svg)](https://pkg.go.dev/github.com/ras0q/goalie)

Goalie (/góʊli/) is a Go library designed to **reliably capture and collect errors from `defer`'d functions**, such as `file.Close()`, `conn.Close()`, or `tx.Rollback()`.

Named for its role, much like a **goalie (goalkeeper)**, Goalie ensures that no errors from deferred cleanup operations are missed at the end of Go function execution.

## Usage

> [!CAUTION]
> Goalie is only for handling errors from `defer`'d functions, not for general error handling.

See [Godoc](https://pkg.go.dev/github.com/ras0q/goalie), [./goalie_test.go](./goalie_test.go) and [./_examples](./_examples) for details.

<!-- Developer note: This sample code is copied from ./_examples/basic/main.go. Keep in sync. -->

```go
package main

import (
    "errors"
    "fmt"
    "os"
    "strconv"

    "github.com/ras0q/goalie"
)

func main() {
    if err := run(); err != nil {
        fmt.Printf("EROOR:\n%+v\n\n", err)

        fmt.Printf("is os.ErrClosed?:     %t\n", errors.Is(err, os.ErrClosed))
        fmt.Printf("is ErrInternal?:      %t\n", errors.Is(err, ErrInternal))
        numError := &strconv.NumError{}
        fmt.Printf("as strconv.NumError?: %t\n", errors.As(err, &numError))
    }

    // Output:
    // ERROR:
    // close go.mod: file already closed
    // internal error: failed to convert string to integer: strconv.Atoi: parsing "N0T 1NTEGER": invalid syntax
    //
    // is os.ErrClosed?:     true
    // is ErrInternal?:      true
    // as strconv.NumError?: true
}

var ErrInternal = errors.New("internal error")

func run() (err error) {
    g := goalie.New()
    defer g.Collect(&err) // ✅ Use g.Collect to collect all captured errors at final.

    // Normal error handling for non-deferred operations should be done separately.
    f, err := os.Open("go.mod")
    if err != nil {
        return fmt.Errorf("%w: failed to open file: %w", ErrInternal, err)
    }
    // defer f.Close()     // 🧐 errcheck: Error return value of `f.Close` is not checked.
    defer g.Guard(f.Close) // ✅ Use g.Guard to capture errors from deferred functions.

    // ❌ This code close the file explicitly by mistake.
    _ = f.Close()

    // ❌ This code always fails.
    _, err = strconv.Atoi("N0T 1NTEGER")
    if err != nil {
        return fmt.Errorf("%w: failed to convert string to integer: %w", ErrInternal, err)
    }

    return nil
}
```

## Integrating Goalie into Existing Projects

For existing projects, we provide a migration tool to automatically insert Goalie's error handling patterns!

The tool will analyze your code and suggest fixes for any `defer` statements that might be missing error handling.

### Usage

> [!CAUTION]
> Always review the changes made by the migrator, especially in complex functions, to ensure correctness.

Run the migrator on your project:

```bash
# Check changes
go run github.com/ras0q/goalie/migrator/cmd/goalie-migrator@latest -diff -fix ./...

# Apply changes
go run github.com/ras0q/goalie/migrator/cmd/goalie-migrator@latest -fix ./...
```

