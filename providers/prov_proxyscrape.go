package providers

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req"
)

// ProxyScrape returns the proxies that can be found on the site https://proxyscrape.com.
func ProxyScrape(proxies *Set, timeout time.Duration) ([]Proxy, error) {
	logger.Debug("providers: Fetching proxies from provider ProxyScrape")

	var wg = &sync.WaitGroup{}
	var links = map[string]string{
		"http":   "https://api.proxyscrape.com/?request=getproxies&proxytype=all&timeout=10000&country=all&ssl=no&anonymity=all",
		"https":  "https://api.proxyscrape.com/?request=getproxies&proxytype=all&timeout=10000&country=all&ssl=yes&anonymity=all",
		"socks4": "https://api.proxyscrape.com/?request=getproxies&proxytype=socks4&timeout=10000&country=all",
		"socks5": "https://api.proxyscrape.com/?request=getproxies&proxytype=socks5&timeout=10000&country=all",
	}

	for ptype, link := range links {
		wg.Add(1)

		go func(ptype, link string) {
			defer wg.Done()

			resp, err := req.Get(link)
			if err != nil {
				logger.Debugf("providers (ProxyScrape): cannot request ProxyScrape API endpoint %v: %v", link, err)
				return
			}

			lines := strings.Split(strings.TrimSpace(resp.String()), "\n")

			for _, rawip := range lines {
				wg.Add(1)

				go func(rawip string) {
					wg.Done()

					rawip = strings.TrimSpace(rawip)
					ip := ptype + "://" + rawip

					if strings.Count(rawip, ".") != 3 {
						logger.Debugf("providers (ProxyScrape): malformed proxy ip %v", rawip)
						return
					}

					components := strings.Split(rawip, ":")
					if len(components) != 2 {
						logger.Debugf("providers (ProxyScrape): invalid proxy ip %v", rawip)
						return
					}

					country, err := countryInfo.FindCountryByIP(components[0])
					if err != nil {
						logger.Debugf("providers (ProxyScrape): cannot find country from ip '%v': %v", components[0], err)
						return
					}

					proxy, err := newProxy(ip, "ProxyScrape", country)
					if err != nil {
						logger.Debugf("providers (ProxyScrape): cannot create new proxy: %v", err)
					}

					proxies.Add(proxy)

				}(rawip)
			}

		}(ptype, link)
	}

	waitTimeout(wg, timeout)

	ps := proxies.All()
	if len(ps) == 0 {
		return ps, fmt.Errorf("providers (ProxyScrape): no proxies could be gathered")
	}

	return ps, nil

}
