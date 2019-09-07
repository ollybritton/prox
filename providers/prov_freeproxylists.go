package providers

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/imroc/req"
)

type freeProxyListsResponse struct {
	Quote struct {
		Table struct {
			Tr []struct {
				Td []string `xml:"td"`
			} `xml:"tr"`
		} `xml:"table"`
	} `xml:"quote"`
}

// findLinks will return the links to resources where the anchor text matches the regex.
func findLinks(url, regex string) (links []string) {
	c := colly.NewCollector()

	c.OnHTML("a", func(e *colly.HTMLElement) {
		regex := regexp.MustCompile(regex)

		if regex.Match([]byte(e.Text)) {
			links = append(links, e.Attr("href"))
		}

	})

	c.Visit(url)

	return links
}

// FreeProxyLists returns the proxies that can be found on the site https://freeproxylists.com
func FreeProxyLists(proxies *Set, timeout time.Duration) ([]Proxy, error) {
	logger.Debug("providers: Fetching proxies from provider FreeProxyLists")

	var wg = &sync.WaitGroup{}
	var lists = []string{
		"http://freeproxylists.com/elite.html",
		"http://freeproxylists.com/anonymous.html",
		"http://freeproxylists.com/https.html",
	}

	for _, list := range lists {
		logger.Debugf("providers (FreeProxyLists): Pulling proxy lists from site %v", list)

		links := findLinks(list, `^detailed list #\d+`)

		for _, link := range links {
			wg.Add(1)

			go func(link string) {
				defer wg.Done()

				components := strings.Split(link, "/")

				if len(components) != 2 {
					logger.Errorf("providers (FreeProxyLists): invalid link type %v", link)
					return
				}

				id := components[1]
				ptype := components[0]

				resource := fmt.Sprintf("http://freeproxylists.com/load_%v_%v", ptype, id)
				resp, err := req.Get(resource)
				if err != nil {
					logger.Debugf("providers (FreeProxyLists): error requesting proxies from site %v: %v", link, err)
					return
				}

				body := resp.String()
				body = strings.ReplaceAll(body, "&lt;", "<")
				body = strings.ReplaceAll(body, "&gt;", ">")

				table := freeProxyListsResponse{}
				err = xml.Unmarshal([]byte(body), &table)

				for _, row := range table.Quote.Table.Tr {
					wg.Add(1)

					go func(info []string) {
						defer wg.Done()

						if len(info) != 6 {
							return
						}

						hasSSL, err := strconv.ParseBool(info[2])
						if err != nil {
							logger.Debugf("providers (FreeProxyLists): invalid boolean '%v' for country information: %v", info[2], err)
							return
						}

						var protocol string

						if hasSSL {
							protocol = "https"
						} else {
							protocol = "http"
						}

						rawip := fmt.Sprintf("%v://%v:%v", protocol, info[0], info[1])

						country, err := countryInfo.FindCountryByName(info[5])
						if err != nil {
							logger.Debugf("providers (FreeProxyLists): cannot deduce ISO country code for country '%v': %v", info[5], err)
							return
						}

						proxy, err := newProxy(rawip, "FreeProxyLists", country)
						if err != nil {
							logger.Debugf("providers (FreeProxyLists): cannot create new proxy: %v", err)
							return
						}
						proxies.Add(proxy)
					}(row.Td)
				}
			}(link)

		}
	}

	waitTimeout(wg, timeout)

	ps := proxies.All()
	if len(ps) == 0 {
		return ps, fmt.Errorf("providers (FreeProxyLists): no proxies could be gathered")
	}

	return ps, nil
}
