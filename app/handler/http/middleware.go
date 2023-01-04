package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Middleware specific for httprouter.Handle.
type Middleware func(httprouter.Handle) httprouter.Handle

// MiddlewareChain is a sequence of middlewares that are able to
// wrap the actual handler.
type MiddlewareChain struct {
	mws []Middleware
}

// NewMiddlewareChain is a constructor for MiddlewareChain.
func NewMiddlewareChain(ms ...Middleware) MiddlewareChain {
	return MiddlewareChain{
		ms,
	}
}

// Chain more Middlewares to existing MiddlewareChain.
func (mc MiddlewareChain) Chain(hs ...Middleware) MiddlewareChain {
	mc.mws = append(mc.mws, hs...)
	return mc
}

// Wrap actual httprouter.Handle with sequence of middlewares.
func (mc MiddlewareChain) Wrap(h httprouter.Handle) httprouter.Handle {
	out := h
	for i := len(mc.mws) - 1; i >= 0; i-- {
		out = mc.mws[i](out)
	}

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		out(w, r, ps)
	}
}

// JSONContentMiddleware set's response content type to application/json.
func JSONContentMiddleware(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		h(w, r, ps)
	}
}
