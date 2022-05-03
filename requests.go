package amizone

import (
	"amizone/internal"
	"amizone/internal/parse"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
)

// doRequest is an internal http request helper to simplify making requests.
// This method takes care of both composing requests, setting custom headers and such as needed.
// If tryLogin is true, the client will attempt to log in if it is not already logged in.
// method must be a valid http request method.
// endpoint must be relative to BaseUrl.
func (a *amizoneClient) doRequest(tryLogin bool, method string, endpoint string, body io.Reader) (*http.Response, error) {
	// Login now if we didn't log in at instantiation.
	if tryLogin && !a.DidLogin() && *a.credentials != (Credentials{}) {
		klog.Infof("doRequest: Attempting to login since we haven't logged in yet.")
		if err := a.login(); err != nil {
			return nil, err
		}
		tryLogin = false // We don't want to attempt another login.
	}

	req, err := http.NewRequest(method, BaseUrl+endpoint, body)
	if err != nil {
		klog.Errorf("%s: %s", ErrFailedToComposeRequest, err)
		return nil, errors.New(ErrFailedToComposeRequest)
	}

	req.Header.Set("User-Agent", internal.Firefox99UserAgent)
	// Amizone uses the referrer to authenticate requests on top of the actual AUTH/session cookies.
	req.Header.Set("Referer", BaseUrl+"/")
	if method == http.MethodPost { // We assume a POST request means submitting a form.
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	response, err := a.client.Do(req)
	if err != nil {
		klog.Errorf("Failed to visit endpoint '%s': %s", endpoint, err)
		return nil, errors.New(fmt.Sprintf("%s: %s", ErrFailedToVisitPage, err))
	}

	// Read the response into a byte array, so we can reuse it.
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return response, errors.New(ErrFailedToReadResponse)
	}
	_ = response.Body.Close()

	response.Body = ioutil.NopCloser(bytes.NewReader(responseBody))

	// If we're directed to try log-ins and the parser determines we're not logged in, we retry.
	if tryLogin && *a.credentials != (Credentials{}) && !parse.LoggedIn(bytes.NewReader(responseBody)) {
		klog.Infof("doRequest: Attempting to login since we're not logged in (likely: session expired).")
		if err := a.login(); err != nil {
			return nil, errors.New(ErrFailedLogin)
		}
		return a.doRequest(false, method, endpoint, body)
	}

	return response, nil
}
