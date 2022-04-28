package amizone

import (
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

func isLoggedIn(client *http.Client) bool {
	jar := client.Jar
	if jar == nil {
		return false
	}

	amizoneUrl, _ := url.Parse(baseUrl)

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

func getNewColly(httpClient *http.Client, loggedIn bool) *colly.Collector {
	return colly.NewCollector(func(c *colly.Collector) {
		c.AllowURLRevisit = true
		c.UserAgent = firefox99UserAgent
		c.AllowedDomains = []string{amizoneDomain}
		c.IgnoreRobotsTxt = true
		c.SetClient(httpClient)

		c.OnResponse(func(r *colly.Response) {
			klog.Infof("Receiving response from amizone with status: %d", r.StatusCode)
		})
		c.OnRequest(func(request *colly.Request) {
			klog.Infof("Sending request to amizone: %s, setting referer...", request.URL.String())
			if loggedIn == true {
				request.Headers.Set("Referer", baseUrl)
			}
		})
	})
}
