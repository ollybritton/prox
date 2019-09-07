package prox

import (
	"strings"
	"time"

	"github.com/ollybritton/prox/providers"
)

// Filter is a function that will either allow or not allow a proxy.
// A filter returns true if a proxy "succeeds", and false if it is not allowed.
type Filter func(p *Proxy) bool

// FilterAllowCountries creates a filter that only allows the countries specified.
// The countries provided must be in the ISO Alpha-2 format (GB, US, etc...)
func FilterAllowCountries(countries []string) Filter {
	logger.Debugf("prox: applying allow countries filter with following countries: %v", countries)
	return func(p *Proxy) bool {
		result := false

		for _, allowedCountry := range countries {
			if allowedCountry == p.Country {
				result = true
				break
			}
		}

		return result
	}
}

// FilterDisallowCountries creates a filter that does not let countries from the
// list to be present.
func FilterDisallowCountries(countries []string) Filter {
	logger.Debugf("prox: applying disallow countries filter with following countries: %v", countries)
	return func(p *Proxy) bool {
		result := true

		for _, allowedCountry := range countries {
			if allowedCountry == p.Country {
				result = false
				break
			}
		}

		return result
	}
}

// FilterProxyTypes creates a filter that only allows specific types of proxies, such as HTTP or SOCKS5.
func FilterProxyTypes(ptypes ...string) Filter {
	logger.Debugf("prox: applying proxy type filter, allowing the following types: %v", ptypes)
	for i := range ptypes {
		ptypes[i] = strings.ToLower(ptypes[i])

		if ptypes[i] != "http" && ptypes[i] != "https" && ptypes[i] != "socks4" && ptypes[i] != "socks5" {
			panic("prox: invalid proxy type specified for filter: " + ptypes[i])
		}
	}

	return func(p *Proxy) bool {
		result := false

		for _, ptype := range ptypes {
			if p.URL.Scheme == ptype {
				result = true
				break
			}
		}

		return result
	}
}

// FilterProxySpeed creates a filter that only allows proxies if they can make a successful
// request in a given timeframe.
func FilterProxySpeed(speed time.Duration) Filter {
	logger.Debugf("prox: applying proxy speed filter with speed of %v", speed)
	return func(p *Proxy) bool {
		return p.CheckSpeed(speed)
	}
}

// FilterProxyConnection creates a filter that will only disallow a proxy if it is not working.
// A timeout of 10 seconds is applied, but if a proxy does timeout it is not marked as not working.
func FilterProxyConnection() Filter {
	logger.Debugf("prox: applying proxy connection filter")
	return func(p *Proxy) bool {
		return p.CheckConnection()
	}
}

// ApplyFilters will apply filters to a list of proxies, and will return a new proxy list.
func ApplyFilters(proxies []providers.Proxy, filters []Filter) []providers.Proxy {
	newProxies := []providers.Proxy{}

	for _, p := range proxies {
		proxy := CastProxy(p)

		// If this is true by the end of the loop, a proxies will be allowed through
		result := true

		for _, filter := range filters {
			if filter(proxy) {
				// Proxy is allowed
				result = true
			} else {
				// Proxy is not allowed
				result = false
				break
			}
		}

		if result == true {
			newProxies = append(newProxies, p)
		}

	}

	return newProxies
}
