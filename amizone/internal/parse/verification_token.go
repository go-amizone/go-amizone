package parse

import (
	"fmt"
	"io"

	"github.com/PuerkitoBio/goquery"
	"k8s.io/klog/v2"
)

const verificationTokenName = "__RequestVerificationToken"

func VerificationToken(body io.Reader) string {
	dom, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		klog.Errorf("failed to parse login page: %s. Was the right page passed?", err.Error())
	}
	return dom.Find(fmt.Sprintf("input[name='%s']", verificationTokenName)).AttrOr("value", "")
}

func VerificationTokenFromDom(dom *goquery.Document) string {
	return dom.Find(fmt.Sprintf("input[name='%s']", verificationTokenName)).AttrOr("value", "")
}
