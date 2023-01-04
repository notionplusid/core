package http

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// GetAuth registers the entity and allows to start watching
// the new tables.
func (h *Handler) GetAuth(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	w.WriteHeader(http.StatusNoContent)
}
