package providers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func proxyScrapeWorker(id int, client *http.Client, jobs chan [2]string, results chan Proxy) {
	for job := range jobs {
		ptype, link := job[0], job[1]

		resp, err := client.Get(link)
		if err != nil {
			logger.Debugf("providers (ProxyScrape): cannot request ProxyScrape API endpoint %v: %v", link, err)
			continue
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Debugf("providers (ProxyScrape): cannot read response body from ProxyScrape")
			continue
		}

		lines := strings.Split(strings.TrimSpace(string(bytes)), "\n")

		for _, rawip := range lines {

			rawip = strings.TrimSpace(rawip)
			ip := ptype + "://" + rawip

			if strings.Count(rawip, ".") != 3 {
				logger.Debugf("providers (ProxyScrape): malformed proxy ip %v", rawip)
				continue
			}

			components := strings.Split(rawip, ":")
			if len(components) != 2 {
				logger.Debugf("providers (ProxyScrape): invalid proxy ip %v", rawip)
				continue
			}

			country, err := countryInfo.FindCountryByIP(components[0])
			if err != nil {
				logger.Debugf("providers (ProxyScrape): cannot find country from ip '%v': %v", components[0], err)
				continue
			}

			proxy, err := newProxy(ip, "ProxyScrape", country)
			if err != nil {
				logger.Debugf("providers (ProxyScrape): cannot create new proxy: %v", err)
			}

			results <- proxy

		}
	}
}

// ProxyScrape returns the proxies that can be found on the site https://proxyscrape.com.
func ProxyScrape(proxies *Set, timeout time.Duration) ([]Proxy, error) {
	logger.Debug("providers: Fetching proxies from provider ProxyScrape")
	client := &http.Client{}

	var links = map[string]string{
		"http":   "https://api.proxyscrape.com/?request=getproxies&proxytype=all&timeout=10000&country=all&ssl=no&anonymity=all",
		"https":  "https://api.proxyscrape.com/?request=getproxies&proxytype=all&timeout=10000&country=all&ssl=yes&anonymity=all",
		"socks4": "https://api.proxyscrape.com/?request=getproxies&proxytype=socks4&timeout=10000&country=all",
		"socks5": "https://api.proxyscrape.com/?request=getproxies&proxytype=socks5&timeout=10000&country=all",
	}

	jobs := make(chan [2]string, 1000)
	results := make(chan Proxy, 1000)

	for i := 0; i < 50; i++ {
		go proxyScrapeWorker(i, client, jobs, results)
	}

	for ptype, link := range links {
		jobs <- [2]string{ptype, link}
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

	ps := proxies.All()
	if len(ps) == 0 {
		return ps, fmt.Errorf("providers (ProxyScrape): no proxies could be gathered")
	}

	return ps, nil

}
