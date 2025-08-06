# 🥅 Goalie

[![Go Reference](https://pkg.go.dev/badge/github.com/ras0q/goalie.svg)](https://pkg.go.dev/github.com/ras0q/goalie)

Goalie (/góʊli/) is a Go library designed to **reliably capture and collect errors from `defer`'d cleanup functions**, such as `file.Close()`, `conn.Close()`, or `tx.Rollback()`.

Named for its role, much like a **goalie (goalkeeper)**, Goalie ensures that no errors from deferred cleanup operations are missed at the end of Go function execution.

## The Story of Common `defer` Mistakes

It's late at night. You've just written some code to read a file and process its contents:

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close() // Oops!
    // ... do something ...
    return nil
}
```

At first glance, this looks fine. But running a static analysis tool like `errcheck` will already warn you:

```
errcheck: Error return value of `f.Close` is not checked
```

Unfortunately, **any error returned by `f.Close()` is silently ignored**. If `f.Close()` fails, you'll never know!

You might try to improve things by logging the error in a deferred function:

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer func() {
        if cerr := f.Close(); cerr != nil {
            // 😶 Only logs the error, does not return it!
            log.Printf("failed to close: %v", cerr)
        }
    }()
    // ... do something ...
    return nil
}
```

But here, while the error is just logged, **the error from `f.Close()` is not returned to the caller** and is effectively lost for upstream error handling.

Worse yet, after being nagged one too many times by errcheck’s warnings about unchecked errors in defer statements, a tired developer might be tempted to silence the error completely. The result? Adding a comment like `//nolint:errcheck` to just make the warning go away—this is the worst possible move:

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    // 😱 Oh no! You can't be serious! Errors are just thrown away!
    defer f.Close() //nolint:errcheck
    // ... do something ...
    return nil
}
```

This practice hides problems instead of solving them, making error detection even harder.

These patterns are surprisingly common—and subtle! Goalie exists to ensure that all errors from deferred cleanup are reliably collected and returned, never lost or forgotten.

## A Better Way: Goalie

With Goalie, you can ensure that all errors from deferred cleanup are properly collected and reported. No more silent failures or hidden cleanup mistakes!

> [!CAUTION]
> Goalie is only for handling errors from cleanup operations, not for general error handling.

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
    defer g.Guard(f.Close) // ✅ Use g.Guard to capture errors from the deferred cleanup function.

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

Run the migrator on your project:

```bash
# Check changes
go run github.com/ras0q/goalie/migrator/cmd/goalie-migrator@latest -diff -fix ./...

# Apply changes
go run github.com/ras0q/goalie/migrator/cmd/goalie-migrator@latest -fix ./...
```

> [!CAUTION]
> Always review the changes made by the migrator, especially in complex functions, to ensure correctness.
