package providers

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
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

func freeProxyListsWorker(id int, client *http.Client, jobs <-chan string, results chan<- Proxy) {
	for link := range jobs {
		components := strings.Split(link, "/")

		if len(components) != 2 {
			logger.Errorf("providers (FreeProxyLists): invalid link type %v", link)
			continue
		}

		id := components[1]
		ptype := components[0]

		resource := fmt.Sprintf("http://freeproxylists.com/load_%v_%v", ptype, id)
		resp, err := client.Get(resource)
		if err != nil {
			logger.Debugf("providers (FreeProxyLists): error requesting proxies from site %v: %v", link, err)
			continue
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Debugf("providers (FreeProxyLists): cannot read response body from ProxyScrape")
			continue
		}

		body := string(bytes)

		body = strings.ReplaceAll(body, "&lt;", "<")
		body = strings.ReplaceAll(body, "&gt;", ">")

		table := freeProxyListsResponse{}

		err = xml.Unmarshal([]byte(body), &table)
		if err != nil && err != io.EOF {
			logger.Errorf("providers (FreeProxyLists): could not unmarshal xml response: %v", err)
			continue
		}

		for _, row := range table.Quote.Table.Tr {

			if len(row.Td) != 6 {
				continue
			}

			hasSSL := row.Td[2] == "true"

			var protocol string

			if hasSSL {
				protocol = "https"
			} else {
				protocol = "http"
			}

			rawip := fmt.Sprintf("%v://%v:%v", protocol, row.Td[0], row.Td[1])

			country, err := countryInfo.FindCountryByName(row.Td[5])
			if err != nil {
				logger.Debugf("providers (FreeProxyLists): cannot deduce ISO country code for country '%v': %v", row.Td[5], err)
				continue
			}

			proxy, err := newProxy(rawip, "FreeProxyLists", country)
			if err != nil {
				logger.Debugf("providers (FreeProxyLists): cannot create new proxy: %v", err)
				continue
			}

			results <- proxy
		}
	}
}

// FreeProxyLists returns the proxies that can be found on the site https://freeproxylists.com
func FreeProxyLists(proxies *Set, timeout time.Duration) ([]Proxy, error) {
	logger.Debug("providers: Fetching proxies from provider FreeProxyLists")
	client := &http.Client{}

	var lists = []string{
		"http://freeproxylists.com/elite.html",
		"http://freeproxylists.com/anonymous.html",
		"http://freeproxylists.com/https.html",
	}

	jobs := make(chan string, 100)
	results := make(chan Proxy, 100)

	for _, list := range lists {
		logger.Debugf("providers (FreeProxyLists): Pulling proxy lists from site %v", list)

		links := findLinks(list, `^detailed list #\d+`)

		for i := 0; i < 50; i++ {
			go freeProxyListsWorker(i, client, jobs, results)
		}

		for _, link := range links {
			jobs <- link
		}
	}
	close(jobs)

	tchan := make(chan struct{})

	go func() {
		for proxy := range results {
			proxies.Add(proxy)
		}

		tchan <- struct{}{}
	}()

	select {
	case <-tchan:
		break
	case <-time.After(timeout):
		break
	}

	ps := proxies.List()
	if len(ps) == 0 {
		return ps, fmt.Errorf("providers (FreeProxyLists): no proxies could be gathered")
	}

	return ps, nil
}
