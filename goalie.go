package goalie

import (
	"errors"
)

type Goalie struct {
	errs []error
}

func New() *Goalie {
	return &Goalie{}
}

func (g *Goalie) Guard(errFn func() error) {
	if err := errFn(); err != nil {
		g.errs = append(g.errs, err)
	}
}

func (g *Goalie) Collect(errp *error) {
	errs := append(g.errs, *errp)
	*errp = errors.Join(errs...)
}
