package goalie

import (
	"errors"
)

type Goalie struct {
	errs           []error
	joinErrorsFunc JoinErrorsFunc
}

func New(options ...Option) *Goalie {
	g := Goalie{
		joinErrorsFunc: errors.Join,
	}

	for _, o := range options {
		o(&g)
	}

	return &g
}

func (g *Goalie) Guard(errFunc func() error) {
	if err := errFunc(); err != nil {
		g.errs = append(g.errs, err)
	}
}

func (g *Goalie) Collect(errp *error) {
	errs := append(g.errs, *errp)
	*errp = g.joinErrorsFunc(errs...)
}

type Option func(*Goalie)

type JoinErrorsFunc func(...error) error

func WithJoinErrorsFunc(joinFunc JoinErrorsFunc) Option {
	return func(g *Goalie) {
		g.joinErrorsFunc = joinFunc
	}
}
