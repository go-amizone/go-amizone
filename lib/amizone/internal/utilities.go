package internal

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"k8s.io/klog/v2"
	"net/http"
	"net/url"
)

type cookieMap map[string]string

func (c cookieMap) contains(key string) bool {
	if _, ok := c[key]; ok {
		return true
	}
	return false
}

// IsLoggedIn returns true if the amizone client has the cookies to be logged in.
// This method does not check if the cookies are still valid.
func IsLoggedIn(client *http.Client) bool {
	jar := client.Jar
	if jar == nil {
		return false
	}

	amizoneUrl, _ := url.Parse("https://" + AmizoneDomain)

	amizoneCookies := func() cookieMap {
		cookieMap := make(cookieMap)
		for _, cookie := range jar.Cookies(amizoneUrl) {
			cookieMap[cookie.Name] = cookie.Value
		}
		return cookieMap
	}()

	for _, key := range []string{".ASPXAUTH", "ASP.NET_SessionId", "__RequestVerificationToken"} {
		if !amizoneCookies.contains(key) {
			return false
		}
	}

	return true
}

// GetNewColly returns a new colly.Collector configured with the http.Client passed and options to allow the client
// to access Amizone inconspicuously. The client is also configured to log all requests and responses.
func GetNewColly(httpClient *http.Client, loggedIn bool) *colly.Collector {
	return colly.NewCollector(func(c *colly.Collector) {
		c.AllowURLRevisit = true
		c.UserAgent = Firefox99UserAgent
		c.AllowedDomains = []string{AmizoneDomain}
		c.IgnoreRobotsTxt = true
		c.SetClient(httpClient)

		c.OnResponse(func(r *colly.Response) {
			klog.Infof("Receiving response from amizone with status: %d", r.StatusCode)
		})
		c.OnRequest(func(request *colly.Request) {
			klog.Infof("Sending request to amizone: %s, setting referer...", request.URL.String())
			if loggedIn == true {
				request.Headers.Set("Referer", fmt.Sprintf("https://%s/", AmizoneDomain))
			}
		})
	})
}
