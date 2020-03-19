// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware is a minimalistic middleware chainer to handle  multiple middleware(s) in correct order.
//
// A Logger, RBAC and JWT middleware are already pre-defined. Check the package documentation for more information.
package middleware

import (
	"net/http"
)

// middleware handler
type middleware func(http.HandlerFunc) http.HandlerFunc

// Chain is holding all added middleware(s).
type Chain struct {
	mws []middleware
}

// New creates an middleware chain.
// It can be empty or multiple mws can be added as argument.
func New(m ...middleware) *Chain {
	return &Chain{append(([]middleware)(nil), m...)}
}

// Add one or more middleware(s) as argument.
func (c *Chain) Add(m ...middleware) *Chain {
	c.mws = append(c.mws, m...)
	return c
}

// All returns the defined middleware(s).
func (c *Chain) All() []middleware {
	return c.mws
}

// Handle all defined middleware(s) in the order they were added to the chain.
func (c *Chain) Handle(h http.HandlerFunc) http.HandlerFunc {
	for i := range c.mws {
		h = c.mws[len(c.mws)-i-1](h)
	}
	return h
}
