package handlers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/ditsuke/go-amizone/amizone"
	"net/http"
)

// authenticatedHandlerWrapper wraps an AuthenticatedHandler, composing a http.HandlerFunc
// This function handles retrieving authentication information from the request, initializing
// an amizone.amizoneClient with the information, and then passing this to the handler.
// This function also handles authentication errors if the auth information is invalid.
func authenticatedHandlerWrapper(c *Cfg, handler AuthenticatedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// We accept credentials through a single custom header, where they're encoded in base64 as "username:password"
		encodedCredentials := r.Header.Get("X-Amizone-Credentials")
		if encodedCredentials == "" {
			w.WriteHeader(http.StatusBadRequest)
		}

		// Here, we attempt to decode the credentials first => then extract the parts
		username, password, err := func() (string, string, error) {
			decoded, err := base64.StdEncoding.DecodeString(encodedCredentials)
			if err != nil {
				return "", "", err
			}
			sepIndex := bytes.IndexRune(decoded, ':')
			if sepIndex == -1 {
				return "", "", errors.New("invalid format")
			}
			return string(decoded[:sepIndex]), string(decoded[sepIndex+1:]), nil
		}()

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
