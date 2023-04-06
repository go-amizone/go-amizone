package parse

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ditsuke/go-amizone/amizone/models"
)

func WifiMacInfo(body io.Reader) (*models.WifiMacInfo, error) {
	dom, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFailedToParse, err)
	}

	macs := make([]net.HardwareAddr, 0, 2)
	nodes := dom.Find("input").FilterFunction(func(_ int, s *goquery.Selection) bool {
		return strings.HasPrefix(s.AttrOr("id", ""), "Mac")
	})
	if nodes.Length() == 0 {
		return nil, errors.New(ErrFailedToParse)
	}

	nodes.Each(func(_ int, s *goquery.Selection) {
		v := s.AttrOr("value", "")
		if v == "" {
			return
		}

		mac, err := net.ParseMAC(v)
		if err != nil {
			// LOG
			return
		}
		macs = append(macs, mac)
	})

	info := models.WifiMacInfo{
		RegisteredAddresses: macs,
		Slots:               nodes.Length(),
		FreeSlots:           nodes.Length() - len(macs),
	}

	info.SetRequestVerificationToken(VerificationTokenFromDom(dom))

	return &info, nil
}
