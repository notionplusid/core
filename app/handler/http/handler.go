package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/notionplusid/core/app/service"
)

// Dep is a dependencies structure that is expected for the Handler to function.
type Dep struct {
	Tenant *service.Tenant
	Table  *service.Table

	IsInternal bool
}

// Validate the Dep.
func (d Dep) Validate() error {
	switch {
	case d.Tenant == nil:
		return errors.New("tenant is required")
	case d.Table == nil:
		return errors.New("table is required")
	}
	return nil
}

// Handler for the HTTP requests.
type Handler struct {
	d  Dep
	hr *httprouter.Router
}

// New Handler constructor.
func New(ctx context.Context, dep Dep) (*Handler, error) {
	if err := dep.Validate(); err != nil {
		return nil, fmt.Errorf("dep: %s", err)
	}

	mw := NewMiddlewareChain(JSONContentMiddleware)

	h := &Handler{
		d:  dep,
		hr: httprouter.New(),
	}

	h.hr.GET("/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Write([]byte("ok")) // nolint: errcheck
	})

	h.hr.GET("/v1/auth", mw.Wrap(h.GetAuth))
	h.hr.GET("/_ah/warmup", func(_ http.ResponseWriter, _ *http.Request, _ httprouter.Params) {})

	return h, nil
}

// ServeHTTP to implement http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.hr.ServeHTTP(w, r)
}
