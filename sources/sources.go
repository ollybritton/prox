package sources

import (
	"net/http"
	"time"
)

// Source represents a source of proxies.
type Source interface {
	Query(Options) ([]Proxy, error)
}

var (
	// ProxyScrape provides access to proxies from ProxyScrape.com
	ProxyScrape Source = proxyScrape{&http.Client{Timeout: 20 * time.Second}}
)
