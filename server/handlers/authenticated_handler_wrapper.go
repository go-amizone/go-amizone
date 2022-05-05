package handlers

import (
	"github.com/ditsuke/go-amizone/amizone"
	"net/http"
)

// authenticatedHandlerWrapper wraps an AuthenticatedHandler, composing a http.HandlerFunc
// This function handles retrieving authentication information from the request, initializing
// an amizone.amizoneClient with the information, and then passing this to the handler.
// This function also handles authentication errors if the auth information is invalid.
func authenticatedHandlerWrapper(c *Cfg, handler AuthenticatedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters for auth
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		client, err := c.A(
			amizone.Credentials{
				Username: username,
				Password: password,
			},
			nil)

		if err != nil {
			if err.Error() == amizone.ErrInvalidCredentials {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			c.L.Error(err, "error creating amizone client")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// At this point, we're authenticated -- we let the authenticated handler take over
		handler(w, r, client)
	}
}
