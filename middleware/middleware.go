// Package middleware is a minimalistic middleware helper for the normal http.Handler and the httprouter.Handle of julienschmidt.
// You can chain multiple middlewares and the get handled in the right order
// See https://github.com/patrickascher/go-middleware for more information and examples.
package middleware

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type Interface interface {
	Add()
	GetAll()
	Handle()
}

// Middleware handler
type Middleware func(http.HandlerFunc) http.HandlerFunc
type MiddlewareJR func(httprouter.Handle) httprouter.Handle

// Chain is holding all added middlewares.
type Chain struct {
	middlewares []Middleware
}

// chain is holding all added middlewares.
type ChainJR struct {
	middlewares []MiddlewareJR
}

// New creates a new chain of middlewares.
func New(m ...Middleware) *Chain {
	return &Chain{append(([]Middleware)(nil), m...)}
}

// NewJR creates a new chain of middlewares for the julienschmidt router.
func NewJR(m ...MiddlewareJR) *ChainJR {
	return &ChainJR{append(([]MiddlewareJR)(nil), m...)}
}

// Add one or more middlewares.
func (c *Chain) Add(m ...Middleware) *Chain {
	c.middlewares = append(c.middlewares, m...)
	return c
}

// Add one or more middlewares.
func (c *ChainJR) Add(m ...MiddlewareJR) *ChainJR {
	c.middlewares = append(c.middlewares, m...)
	return c
}

//GetAll configured middlewares.
func (c *Chain) GetAll() []Middleware {
	return c.middlewares
}

//GetAll configured middlewares.
func (c *ChainJR) GetAll() []MiddlewareJR {
	return c.middlewares
}

//Handle all middlewares in the order they were added to the chain.
func (c *Chain) Handle(h http.HandlerFunc) http.HandlerFunc {
	for i := range c.middlewares {
		h = c.middlewares[len(c.middlewares)-i-1](h)
	}
	return h
}

//Handle all middlewares in the order they were added to the chain.
func (c *ChainJR) Handle(h httprouter.Handle) httprouter.Handle {
	for i := range c.middlewares {
		h = c.middlewares[len(c.middlewares)-i-1](h)
	}
	return h
}
