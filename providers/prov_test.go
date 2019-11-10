package providers_test

import (
	"testing"
	"time"

	"github.com/ollybritton/prox/providers"
)

func testProvider(name string, provider providers.Provider) func(*testing.T) {
	return func(t *testing.T) {
		proxies := providers.NewSet()
		ps, err := provider(proxies, 10*time.Second)

		if err != nil {
			t.Errorf("providers (%v): error occurred when scraping proxies: %v", name, err)
		} else {
			t.Logf("providers (%v): found %d proxies", name, len(ps))
		}
	}
}

func TestProviders(t *testing.T) {
	t.Run("FreeProxyLists", testProvider("FreeProxyLists", providers.FreeProxyLists))
	t.Run("ProxyScrape", testProvider("ProxyScrape", providers.ProxyScrape))
	t.Run("Static", testProvider("Static", providers.Static))
}
