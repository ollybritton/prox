package providers

import (
	"net/url"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Provider is a type alias representing a proxy provider.
type Provider func(*Set, time.Duration) ([]Proxy, error)

// Proxy represents a proxy
type Proxy struct {
	URL      *url.URL `json:"url"`
	Provider string   `json:"providers"`
	Country  string   `json:"country"`

	Used bool
}

func newProxy(rawip string, provider string, country string) (Proxy, error) {
	u, err := url.Parse(rawip)
	if err != nil {
		return Proxy{}, errors.Wrap(err, "providers: cannot parse ip to URL")
	}

	return Proxy{
		URL:      u,
		Provider: provider,
		Country:  country,
	}, nil
}

// Set is a utility for storing the proxies in a concurrency-safe way
type Set struct {
	m       sync.Mutex
	proxies map[Proxy]bool
}

// Add adds a new proxy to the set.
func (s *Set) Add(p Proxy) {
	s.m.Lock()

	exists := s.proxies[p]
	if !exists {
		s.proxies[p] = true
	}

	s.m.Unlock()
}

// In checks wheter a proxy is in the set.
func (s *Set) In(p Proxy) bool {
	s.m.Lock()
	m := s.proxies[p]
	s.m.Unlock()

	return m
}

// List returns all the proxies in the set as a slice.
func (s *Set) List() (proxies []Proxy) {
	s.m.Lock()

	keys := make([]Proxy, 0, len(s.proxies))
	for k := range s.proxies {
		keys = append(keys, k)
	}

	s.m.Unlock()

	return keys
}

// All returns all the proxies in the set as a map.
func (s *Set) All() (proxies map[Proxy]bool) {
	s.m.Lock()
	defer s.m.Unlock()

	return s.proxies
}

// Remove removes a proxy from a set.
// If the proxy doesn't exist, no change is made.
func (s *Set) Remove(proxy Proxy) {
	s.m.Lock()

	delete(s.proxies, proxy)
	s.m.Unlock()
}

// Random gets a random proxy from the set.
func (s *Set) Random() Proxy {
	s.m.Lock()
	defer s.m.Unlock()

	for k := range s.proxies {
		return k
	}

	return Proxy{}
}

// FromCountries gets a random proxy from the specified countries.
func (s *Set) FromCountries(countries []string) Proxy {
	s.m.Lock()
	defer s.m.Unlock()

	for k := range s.proxies {
		for _, c := range countries {
			if k.Country == c {
				return k
			}
		}
	}

	return Proxy{}
}

// Length gets the amount of proxies being stores.
func (s *Set) Length() int {
	s.m.Lock()
	size := len(s.proxies)
	s.m.Unlock()

	return size
}

// NewSet creates a new set.
func NewSet() *Set {
	return &Set{
		proxies: make(map[Proxy]bool),
	}
}
