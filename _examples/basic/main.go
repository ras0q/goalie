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
