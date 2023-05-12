package amizone

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ditsuke/go-amizone/amizone/internal"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	"k8s.io/klog/v2"
)

const (
	ErrNon200StatusCode = "received non-200 status code from amizone - is it down?"
)

// doRequest is an internal http request helper to simplify making requests.
// This method takes care of both composing requests, setting custom headers and such as needed.
// If tryLogin is true, the Client will attempt to log in if it is not already logged in.
// method must be a valid http request method.
// endpoint must be relative to BaseUrl.
func (a *Client) doRequest(tryLogin bool, method string, endpoint string, body io.Reader) (*http.Response, error) {
	if *a.credentials == (Credentials{}) {
		return nil, fmt.Errorf("%s: invalid credentials", ErrFailedLogin)
	}

	// Login now if we didn't log in at instantiation.
	if tryLogin && !a.DidLogin() {
		klog.Infof("doRequest: Attempting to login since we haven't logged in yet.")
		if err := a.login(); err != nil {
			return nil, err
		}
		tryLogin = false // We don't want to attempt another login.
	}

	req, err := http.NewRequest(method, BaseURL+endpoint, body)
	if err != nil {
		klog.Errorf("%s: %s", ErrFailedToComposeRequest, err)
		return nil, errors.New(ErrFailedToComposeRequest)
	}

	req.Header.Set("User-Agent", internal.Firefox99UserAgent)
	// Amizone uses the referrer to authenticate requests on top of the actual AUTH/session cookies.
	req.Header.Set("Referer", BaseURL+"/")
	if method == http.MethodPost { // We assume a POST request means submitting a form.
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// TODO: check error handling logic following here
	response, err := a.httpClient.Do(req)
	if err != nil {
		klog.Errorf("Failed to visit endpoint '%s': %s", endpoint, err)
		return nil, fmt.Errorf("%s: %w", ErrFailedToVisitPage, err)
	}

	// Amizone uses code 200 even for POST requests, so we make sure we have that before proceeding.
	if response.StatusCode != http.StatusOK {
		klog.Warningf("Received non-200 status code from endpoint '%s': %d. Amizone down?", endpoint, response.StatusCode)
		return nil, fmt.Errorf("%s: %d", ErrNon200StatusCode, response.StatusCode)
	}

	// Read the response into a byte array, so we can reuse it.
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return response, errors.New(ErrFailedToReadResponse)
	}
	_ = response.Body.Close()

	response.Body = io.NopCloser(bytes.NewReader(responseBody))

	// If we're directed to try logging-in and the parser determines we're not, we retry.
	if tryLogin && *a.credentials != (Credentials{}) && !parse.IsLoggedIn(bytes.NewReader(responseBody)) {
		klog.Infof("doRequest: Attempting to login since we're not logged in (likely: session expired).")
		if err := a.login(); err != nil {
			return nil, errors.New(ErrFailedLogin)
		}
		return a.doRequest(false, method, endpoint, body)
	}

	return response, nil
}
