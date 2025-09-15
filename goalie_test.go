package goalie_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/ras0q/goalie"
)

var (
	errInternal = errors.New("internal error")
)

// This function has too many bugs 😅
func countLines(path string) (_ int, err error) {
	g := goalie.New()
	defer g.Collect(&err)

	f, err := os.Open(path)
	if err != nil {
		return -1, err
	}
	defer g.Guard(f.Close)

	// This code is incorrect.
	// Since it closes with defer, you must not close it explicitly.
	if err := f.Close(); err != nil {
		return -1, err
	}

	// This function is always fails.
	return -1, errInternal
}

// assert helper
func assert[T comparable](t *testing.T, expected, got T) {
	t.Helper()
	if got != expected {
		t.Fatalf("assertion failed (expected: %+v, got: %+v)", expected, got)
	}
}

func Test_Goalie(t *testing.T) {
	type testcase struct {
		path                string
		isFileNotExistError bool
		isFileClosedError   bool
		isInternalError     bool
	}

	run := func(t *testing.T, tc testcase) {
		t.Helper()

		n, err := countLines(tc.path)
		if err != nil {
			t.Logf("\nerr:\n%+v", err)
			assert(t, tc.isFileNotExistError, errors.Is(err, os.ErrNotExist))
			assert(t, tc.isFileClosedError, errors.Is(err, os.ErrClosed))
			assert(t, tc.isInternalError, errors.Is(err, errInternal))
			return
		}
		assert(t, true, n > 0)
	}

	testcases := map[string]testcase{
		"capture both deferred file-closing error and normal error": {
			path:                "goalie_test.go",
			isFileNotExistError: false,
			isFileClosedError:   true,
			isInternalError:     true,
		},
		"return file not found error before setting defer": {
			path:                "nonexistent.txt",
			isFileNotExistError: true,
			isFileClosedError:   false,
			isInternalError:     false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Custom error wrapper for testing.
// This does not implement Unwrap, so [errors.Is] and [errors.As] do not work against wrapped errors.
type customWrapError struct {
	err error
}

func (e *customWrapError) Error() string {
	return fmt.Sprintf("custom error: %v", e.err)
}

func customWrapper(err error) error {
	return &customWrapError{err: err}
}

func Test_SetDefaultWrapErrorFunc(t *testing.T) {
	type testcase struct {
		wrapErrorFunc goalie.WrapErrorFunc
		errAssertFunc func(t *testing.T, err error)
	}

	f := func() (err error) {
		g := goalie.New()
		defer g.Collect(&err)

		defer g.Guard(func() error {
			return os.ErrClosed
		})

		return nil
	}

	testcases := map[string]testcase{
		"no wrapping": {
			wrapErrorFunc: nil,
			errAssertFunc: func(t *testing.T, err error) {
				assert(t, true, errors.Is(err, os.ErrClosed))
			},
		},
		"wrap with custom error": {
			wrapErrorFunc: customWrapper,
			errAssertFunc: func(t *testing.T, err error) {
				assert(t, false, errors.Is(err, os.ErrClosed)) // customWrapError does not implement Unwrap, so errors.Is returns false
				assert(t, true, errors.As(err, new(*customWrapError)))
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			goalie.SetDefaultWrapErrorFunc(tc.wrapErrorFunc)
			t.Cleanup(func() { goalie.SetDefaultWrapErrorFunc(nil) })

			err := f()
			tc.errAssertFunc(t, err)
		})
	}
}
