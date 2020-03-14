package prox

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ollybritton/prox/providers"
)

// Provider is a wrapper around the providers.Provider type, giving information about the provider along with the actual
// provider function itself.
type Provider struct {
	Name             string
	InternalProvider providers.Provider
}

// FreeProxyLists defines the 'FreeProxyLists' provider.
var FreeProxyLists = Provider{"FreeProxyLists", providers.FreeProxyLists}

// ProxyScrape defines the 'ProxyScrape' provider.
var ProxyScrape = Provider{"ProxyScrape", providers.ProxyScrape}

// GetProxyList defines the 'GetProxyList' provider.
var GetProxyList = Provider{"GetProxyList", providers.GetProxyList}

// Static defines the 'Static' provider.
var Static = Provider{"Static", providers.Static}

// Providers is a global variable which allows translation between the names of providers
// and the provider functions themselves.
var Providers = map[string]Provider{
	"FreeProxyLists": FreeProxyLists,
	"ProxyScrape":    ProxyScrape,
	"GetProxyList":   GetProxyList,
	"Static":         Static,
}

// panicValidProvider will panic if the string specified does not correspond to a valid provider.
func panicValidProvider(providerName string) {
	if Providers[providerName].InternalProvider == nil {
		panic(errors.New("invalid provider type " + providerName))
	}
}

// GetProvider gets the provider by name.
func GetProvider(providerName string) Provider {
	panicValidProvider(providerName)
	return Providers[providerName]
}

// GetProviders gets multiple providers by name.
func GetProviders(providerNames ...string) []Provider {
	results := []Provider{}

	for _, providerName := range providerNames {
		panicValidProvider(providerName)
		results = append(results, Providers[providerName])
	}

	return results
}

// MultiProvider creates a new hybrid-provider from a set of existing ones.
// It will fetch all the proxies from all providers asynchronously.
func MultiProvider(givenProviders ...Provider) Provider {
	names := []string{}

	for _, provider := range givenProviders {
		names = append(names, provider.Name)
	}

	name := fmt.Sprintf("Multi{%v}", strings.Join(names, "|"))

	return Provider{name, func(proxies *providers.Set, timeout time.Duration) ([]providers.Proxy, error) {
		var wg = &sync.WaitGroup{}

		for _, provider := range givenProviders {
			wg.Add(1)

			go func(provider providers.Provider) {
				provider(proxies, timeout)
				wg.Done()
			}(provider.InternalProvider)
		}

		waitTimeout(wg, timeout)

		ps := proxies.List()
		if len(ps) == 0 {
			return ps, fmt.Errorf("providers (%v): no proxies could be gathered", name)
		}

		return ps, nil
	}}
}

// FreezeProvider will gather proxies from the provider given one last time
// and then use those instead of new ones.
func FreezeProvider(providerName string, timeout time.Duration) providers.Provider {
	panicValidProvider(providerName)

	ps, err := Providers[providerName].InternalProvider(providers.NewSet(), timeout)
	return func(proxies *providers.Set, timeout time.Duration) ([]providers.Proxy, error) {
		if err != nil {
			return []providers.Proxy{}, err
		}

		return ps, nil
	}
}
