package main

import (
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/ras0q/goalie"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %+v\n", err)
	}
}

func run() (err error) {
	g := goalie.New(
		goalie.WithJoinErrorsFunc(errors.Join),
	)
	// At the end of the function, Goalie collects all captured errors
	defer g.Collect(&err)

	// CAUTION: Normal error handling for non-deferred operations.
	f, err := os.Open("main.go")
	if err != nil {
		return errors.Errorf("failed to open file: %w", err)
	}

	// Use Goalie.Guard for errors from `defer`'d functions (e.g. closing a file)
	defer g.Guard(f.Close)

	_ = f.Close()

	// Simulate another `defer`'d cleanup operation that might return an error
	defer g.Guard(func() error {
		return errors.New("error from a deferred cleanup operation")
	})

	fmt.Println("Operations successful (simulated).")

	return nil
}
