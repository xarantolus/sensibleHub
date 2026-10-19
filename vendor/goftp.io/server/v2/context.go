// Copyright 2020 The goftp Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"time"
)

// Context represents a context the driver may want to know
type Context struct {
	ctx   context.Context
	Sess  *Session
	Cmd   string                 // request command on this request
	Param string                 // request param on this request
	Data  map[string]interface{} // share data between middlewares
}

var _ context.Context = (*Context)(nil)

func newContext(parent context.Context, sess *Session, cmd, param string, data map[string]interface{}) *Context {
	return &Context{
		ctx:   parent,
		Sess:  sess,
		Cmd:   cmd,
		Param: param,
		Data:  data,
	}
}

// Deadline returns the deadline associated with this context, if any.
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

// Done returns a channel that is closed when the work done on behalf of this context is complete.
func (c *Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

// Err returns the error, if any, that caused the context to be canceled.
func (c *Context) Err() error {
	return c.ctx.Err()
}

// Value returns the value associated with this context.
func (c *Context) Value(key any) any {
	return c.ctx.Value(key)
}
