//go:generate go-bindata -pkg $GOPACKAGE data/

package providers

import (
	"fmt"
	"strings"
	"time"
)

// Static provides access to a static proxy list that can be used offline.
func Static(proxies *Set, timeout time.Duration) ([]Proxy, error) {
	bytes, err := Asset("data/proxies.txt")
	if err != nil {
		return []Proxy{}, err
	}

	for _, info := range strings.Split(string(bytes), "\n") {
		func(info string) {

			components := strings.Split(info, " ")
			if len(components) != 2 {
				logger.Debugf("providers (Static): invalid static proxy format %v", info)
				return
			}

			url := components[0]
			country := components[1]

			proxy, err := newProxy(url, "Static", country)
			if err != nil {
				logger.Debugf("providers (Static): cannot create new proxy: %v", err)
				return
			}

			proxies.Add(proxy)

		}(info)
	}

	ps := proxies.List()
	if len(ps) == 0 {
		return ps, fmt.Errorf("providers (Static): no proxies could be gathered")
	}

	return ps, nil
}
