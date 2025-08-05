# 🥅 Goalie

[![Go Reference](https://pkg.go.dev/badge/github.com/ras0q/goalie.svg)](https://pkg.go.dev/github.com/ras0q/goalie)

Goalie (/góʊli/) is a Go library designed to **reliably capture and collect errors from `defer`'d functions**, such as `file.Close()`, `conn.Close()`, or `tx.Rollback()`.

Named for its role, much like a **goalie (goalkeeper)**, Goalie ensures that no errors from deferred cleanup operations are missed at the end of Go function execution.

## Usage

> [!CAUTION]
> Goalie is only for handling errors from `defer`'d functions, not for general error handling.

See [Godoc](https://pkg.go.dev/github.com/ras0q/goalie), [./goalie_test.go](./goalie_test.go) and [./_examples](./_examples) for details.

```go
package main

import (
    "errors"
    "fmt"
    "os"

    "github.com/ras0q/goalie"
)

func main() {
    if err := run(); err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}

func run() (err error) {
    g := goalie.New()
    // Collects all captured errors at the end of the function,
    // ensuring that deferred errors are propagated.
    defer g.Collect(&err)

    // Normal error handling for non-deferred operations should be done separately.
    f, err := os.Open("example.txt")
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }

    // Use g.Guard to capture errors from `defer`'d functions (e.g., file.Close(), conn.Close()).
    defer g.Guard(f.Close)

    // Simulate another `defer`'d cleanup operation that might return an error
    defer g.Guard(func() error {
        return errors.New("error from a deferred cleanup operation")
    })

    return nil
}
```

## Integrating Goalie into Existing Projects

For existing projects, we provide a migration tool to automatically insert Goalie's error handling patterns!

The tool will analyze your code and suggest fixes for any `defer` statements that might be missing error handling.

### Usage

Run the migrator on your project:

```bash
# Check changes
go run github.com/ras0q/goalie/migrator/cmd/goalie-migrator@latest -diff -fix ./...

# Apply changes
go run github.com/ras0q/goalie/migrator/cmd/goalie-migrator@latest -fix ./...
```

> [!CAUTION]
> Always review the changes made by the migrator, especially in complex functions, to ensure correctness.
