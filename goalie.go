// Goalie is a Go library designed to reliably capture and collect errors from `defer`'d functions,
// such as `file.Close()`, `conn.Close()`, or `tx.Rollback()`.
package goalie

import (
	"errors"
)

// Goalie is the main struct that manages captured error.
type Goalie struct {
	errs           []error
	joinErrorsFunc JoinErrorsFunc
}

// New creates a new Goalie instance.
func New(options ...Option) *Goalie {
	g := Goalie{
		joinErrorsFunc: errors.Join,
	}

	for _, o := range options {
		o(&g)
	}

	return &g
}

// Collect captures all errors collected by Goalie and joins them into a single error,
// assigning it to `errp` (a pointer to the function's return error variable).
//
// Use this method in a `defer` statement at the top of a function to ensure
// all errors are collected and propagated before the function returns.
//
// Example:
//
//	func doSomething() (err error) {
//		g := New()
//		defer g.Collect(&err)
//
//		// ... operations that might use g.Guard ...
//
//		return nil
//	}
func (g *Goalie) Collect(errp *error) {
	errs := append(g.errs, *errp)
	*errp = g.joinErrorsFunc(errs...)
}

// Guard executes the given function `errFunc` and captures any error returned.
//
// This is useful for capturing errors from `defer`'d functions that do not return an error to the caller.
//
// Example:
//
//	file, _ := os.Open("somefile.txt")
//	defer g.Guard(file.Close)
func (g *Goalie) Guard(errFunc func() error) {
	if err := errFunc(); err != nil {
		g.errs = append(g.errs, err)
	}
}

// Option is a function that configures a [Goalie] instance.
type Option func(*Goalie)

// JoinErrorsFunc is a function type for joining multiple errors into a single error.
// By default, Goalie uses [errors.Join].
type JoinErrorsFunc func(...error) error

// WithJoinErrorsFunc sets the function used to join errors.
//
// This option allows you to customize how captured errors are combined.
// For example, you can use a custom function like [github.com/cockroachdb/errors.Join]
// to include stack traces or other custom error wrapping logic.
func WithJoinErrorsFunc(joinFunc JoinErrorsFunc) Option {
	return func(g *Goalie) {
		g.joinErrorsFunc = joinFunc
	}
}
