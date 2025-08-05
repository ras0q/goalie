// nolint: errcheck
package main

import (
	"fmt"
)

func main() {
	if _, err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func run() (int, error) {
	defer f()        // want `missed error in defer statement: f\(\)`
	defer g("hello") // want `missed error in defer statement: g\(\"hello\"\)`

	s := S{}
	defer s.f()        // want `missed error in defer statement: s\.f\(\)`
	defer s.g("world") // want `missed error in defer statement: s\.g\(\"world\"\)`

	return 0, nil
}

func f() error {
	return fmt.Errorf("error from f()")
}

func g(s string) error {
	return fmt.Errorf("error from g(%s)", s)
}

type S struct{}

func (S) f() error {
	return fmt.Errorf("error from f()")
}

func (S) g(s string) error {
	return fmt.Errorf("error from g(%s)", s)
}
