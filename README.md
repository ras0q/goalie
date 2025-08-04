# 🥅 Goalie

[![Go Reference](https://pkg.go.dev/badge/github.com/ras0q/goalie.svg)](https://pkg.go.dev/github.com/ras0q/goalie)

Goalie (/góʊli/) is a Go library for **reliably capturing errors from `defer`'d functions** like `file.Close()`, `conn.Close()`, or `tx.Rollback()`.

It collects and returns these errors from `defer`'d functions. They are never missed!

The name "Goalie" comes from its role in catching errors at the end in Go.

## Usage

> [!CAUTION]
> Goalie is only for handling errors from `defer`'d functions, not for general error handling.

See [./goalie_test.go](./goalie_test.go) and [./_examples](./_examples) for details.

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
    // At the end of the function, Goalie collects all captured errors
	defer g.Collect(&err)

	// CAUTION: Normal error handling for non-deferred operations.
	f, err := os.Open("example.txt")
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	// Use Goalie.Guard for errors from `defer`'d functions (e.g. closing a file)
	defer g.Guard(f.Close)

	// Simulate another `defer`'d cleanup operation that might return an error
	defer g.Guard(func() error {
		return errors.New("error from a deferred cleanup operation")
	})

	fmt.Println("Operations successful (simulated).")

	return nil
}
```
