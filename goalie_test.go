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

func Test_SetFallbackWrapErrorFunc(t *testing.T) {
	errUnexpected := errors.New("unexpected error")
	type testcase struct {
		wrapErrorFunc       goalie.WrapErrorFunc
		path                string
		isFileNotExistError bool
		isFileClosedError   bool
		isInternalError     bool
		isUnexpectedError   bool
	}

	run := func(t *testing.T, tc testcase) {
		t.Helper()

		goalie.SetFallbackWrapErrorFunc(tc.wrapErrorFunc)
		t.Cleanup(func() { goalie.SetFallbackWrapErrorFunc(nil) })

		_, err := countLines(tc.path)
		if err != nil {
			assert(t, tc.isFileNotExistError, errors.Is(err, os.ErrNotExist))
			assert(t, tc.isFileClosedError, errors.Is(err, os.ErrClosed))
			assert(t, tc.isInternalError, errors.Is(err, errInternal))
			assert(t, tc.isUnexpectedError, errors.Is(err, errUnexpected))
			return
		}
	}

	testcases := map[string]testcase{
		"no wrapping": {
			wrapErrorFunc:       nil,
			path:                "goalie_test.go",
			isFileNotExistError: false,
			isFileClosedError:   true,
			isInternalError:     true,
			isUnexpectedError:   false,
		},
		"wrap with custom error": {
			wrapErrorFunc:       func(err error) error { return fmt.Errorf("%w, %w", errUnexpected, err) },
			path:                "goalie_test.go",
			isFileNotExistError: false,
			isFileClosedError:   true,
			isInternalError:     true,
			isUnexpectedError:   true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func Test_SetFallbackJoinErrorsFunc(t *testing.T) {
	errUnexpected := errors.New("unexpected error")
	type testcase struct {
		joinErrorsFunc      goalie.JoinErrorsFunc
		path                string
		isFileNotExistError bool
		isFileClosedError   bool
		isInternalError     bool
		isUnexpectedError   bool
	}

	run := func(t *testing.T, tc testcase) {
		t.Helper()

		goalie.SetFallbackJoinErrorsFunc(tc.joinErrorsFunc)
		t.Cleanup(func() { goalie.SetFallbackJoinErrorsFunc(nil) })

		_, err := countLines(tc.path)
		if err != nil {
			assert(t, tc.isFileNotExistError, errors.Is(err, os.ErrNotExist))
			assert(t, tc.isFileClosedError, errors.Is(err, os.ErrClosed))
			assert(t, tc.isInternalError, errors.Is(err, errInternal))
			assert(t, tc.isUnexpectedError, errors.Is(err, errUnexpected))
			return
		}
	}

	testcases := map[string]testcase{
		"no custom join": {
			joinErrorsFunc:      nil,
			path:                "goalie_test.go",
			isFileNotExistError: false,
			isFileClosedError:   true,
			isInternalError:     true,
			isUnexpectedError:   false,
		},
		"join with custom error": {
			joinErrorsFunc:      func(err ...error) error { return errors.Join(append([]error{errUnexpected}, err...)...) },
			path:                "goalie_test.go",
			isFileNotExistError: false,
			isFileClosedError:   true,
			isInternalError:     true,
			isUnexpectedError:   true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
