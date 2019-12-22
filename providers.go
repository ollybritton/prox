package prox

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ollybritton/prox/providers"
)

// Providers is a global variable which allows translation between the names of providers
// and the provider functions themselves.
var Providers = map[string]providers.Provider{
	"FreeProxyLists": providers.FreeProxyLists,
	"ProxyScrape":    providers.ProxyScrape,
	"Static":         providers.Static,
}

// panicValidProvider will panic if the string specified does not correspond to a valid provider.
func panicValidProvider(providerName string) {
	if Providers[providerName] == nil {
		panic(errors.New("invalid provider type " + providerName))
	}
}

// GetProvider gets the provider by name.
func GetProvider(providerName string) providers.Provider {
	panicValidProvider(providerName)
	return Providers[providerName]
}

// MultiProvider creates a new hybrid-provider from a set of existing ones.
// It will fetch all the proxies from all providers asynchronously.
func MultiProvider(providerNames ...string) providers.Provider {
	for _, providerName := range providerNames {
		panicValidProvider(providerName)
	}

	name := fmt.Sprintf("Multi{%v}", strings.Join(providerNames, "|"))

	return func(proxies *providers.Set, timeout time.Duration) ([]providers.Proxy, error) {
		var wg = &sync.WaitGroup{}

		for _, providerName := range providerNames {
			wg.Add(1)

			go func(provider providers.Provider) {
				provider(proxies, timeout)
				wg.Done()
			}(Providers[providerName])
		}

		waitTimeout(wg, timeout)

		ps := proxies.List()
		if len(ps) == 0 {
			return ps, fmt.Errorf("providers (%v): no proxies could be gathered", name)
		}

		return ps, nil
	}
}

// FreezeProvider will gather proxies from the provider given one last time
// and then use those instead of new ones.
func FreezeProvider(providerName string, timeout time.Duration) providers.Provider {
	panicValidProvider(providerName)

	ps, err := Providers[providerName](providers.NewSet(), timeout)
	return func(proxies *providers.Set, timeout time.Duration) ([]providers.Proxy, error) {
		if err != nil {
			return []providers.Proxy{}, err
		}

		return ps, nil
	}
}
