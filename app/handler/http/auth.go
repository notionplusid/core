package http

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// GetAuth registers the entity and allows to start watching
// the new tables.
func (h *Handler) GetAuth(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if h.d.IsInternal {
		WriteHTTPErr(w, http.StatusGone, NewHTTPErr(
			HTTPErrCodeGone,
			"Application is running in `internal` mode",
			"This endpoint is meant to be used only for the public setup",
		))
		return
	}

	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		WriteHTTPErr(w, http.StatusUnprocessableEntity, NewHTTPErr(
			HTTPErrCodeNoAuthCode,
			"`code` is required for the workspace to be registered",
			"Something is wrong with the Notion redirect?",
		))
		return
	}

	ws, err := h.d.Tenant.AuthWorkspace(r.Context(), code)
	if err != nil {
		log.Printf("HTTP: Auth: couldn't authorise a workspace: %s", err)
		WriteInternalServerErr(w)
		return
	}

	_, err = h.d.Tenant.RegisterWorkspace(r.Context(), ws)
	if err != nil {
		log.Printf("HTTP: Auth: couldn't register a workspace: %s", err)
		WriteInternalServerErr(w)
		return
	}

	http.RedirectHandler("https://notionplusid.app/welcome", http.StatusFound).
		ServeHTTP(w, r)
}
