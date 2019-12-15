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
	m          sync.Mutex
	membership map[string]bool
	proxies    []Proxy
}

// Add adds a new proxy to the set.
func (s *Set) Add(p Proxy) {
	s.m.Lock()

	exists := s.membership[p.URL.Host]
	if !exists {
		s.membership[p.URL.Host] = true
		s.proxies = append(s.proxies, p)
	}

	s.m.Unlock()
}

// In checks wheter a proxy is in the set.
func (s *Set) In(p Proxy) bool {
	s.m.Lock()
	m := s.membership[p.URL.Host]
	s.m.Unlock()

	return m
}

// All returns all the proxies in the set at the current moment.
func (s *Set) All() (proxies []Proxy) {
	s.m.Lock()
	proxies = s.proxies
	s.m.Unlock()

	return proxies
}

// Remove removes a proxy from a set.
// If the proxy doesn't exist, no change is made.
func (s *Set) Remove(proxy Proxy) {
	s.m.Lock()

	for i, p := range s.proxies {
		if p.URL.String() == proxy.URL.String() {
			s.proxies[len(s.proxies)-1], s.proxies[i] = s.proxies[i], s.proxies[len(s.proxies)-1]
			s.proxies = s.proxies[:len(s.proxies)-1]
			break
		}
	}

	delete(s.membership, proxy.URL.Host)
	s.m.Unlock()
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
		membership: make(map[string]bool),
	}
}
