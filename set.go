package prox

import (
	"errors"
	"math/rand"
	"sync"
)

var (
	// ErrNoneAvaliable happens when the user asks for a proxy with specific critera but they can't be met.
	// E.g. if a user asks for a proxy from France, but none exist.
	// E.g. if a user asks for a proxy from an empty set.
	ErrNoneAvaliable = errors.New("no proxies with that critera are avaliable")
)

// Set is a utility for storing the proxies in a concurrency-safe way and accessing stored proxies fast.
// It organises proxies into countries so that proxies from specific locations can be found without searching through all
// of the proxies.
type Set struct {
	m       sync.Mutex
	proxies map[string]map[Proxy]bool
}

// Add adds a proxy to the set. If the proxy is already in the set, nothing changes.
func (s *Set) Add(p Proxy) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.proxies[p.Country] == nil {
		s.proxies[p.Country] = make(map[Proxy]bool)
	}

	s.proxies[p.Country][p] = true
}

// Remove removes a proxy from the set. If the proxy already exists, then nothing happens.
func (s *Set) Remove(p Proxy) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.proxies[p.Country] == nil {
		return
	}

	delete(s.proxies[p.Country], p)
}

// In checks if a proxy is contained within the set.
func (s *Set) In(p Proxy) bool {
	s.m.Lock()
	defer s.m.Unlock()

	if s.proxies[p.Country] == nil {
		return false
	}

	return s.proxies[p.Country][p]
}

// Array returns all of the proxies as an array of type []Proxy.
func (s *Set) Array() []Proxy {
	s.m.Lock()
	defer s.m.Unlock()

	arr := []Proxy{}

	for _, proxyMap := range s.proxies {
		for proxy := range proxyMap {
			arr = append(arr, proxy)
		}
	}

	return arr
}

// GetBiasedRandom gets a proxy in a biasedly-random way. This method is useful if you don't care about an even distribution
// and just need some proxy. It is biased because it first selects a random country and then a random proxy from inside that
// country. This is a consequence of the data structure used. It also relies on Go's random ordering of map keys which can
// add a bias too.
func (s *Set) GetBiasedRandom() (Proxy, error) {
	s.m.Lock()
	defer s.m.Unlock()

	for _, proxySet := range s.proxies {
		for proxy := range proxySet {
			return proxy, nil
		}
	}

	// This only occurs if there are no proxies.
	return Proxy{}, ErrNoneAvaliable
}

// GetRandom gets a proxy in a random way, giving each proxy an equal chance of being selected unlike GetBiasedRandom. It is
// however more expensive as the set needs to be converted into an array which could be more intensive.
func (s *Set) GetRandom() (Proxy, error) {
	arr := s.Array()

	if len(arr) == 0 {
		return Proxy{}, ErrNoneAvaliable
	}

	return arr[rand.Intn(len(arr))], nil
}

// GetRandomFromCountry gets a random proxy from a given country.
func (s *Set) GetRandomFromCountry(country string) (Proxy, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.proxies[country] == nil {
		return Proxy{}, ErrNoneAvaliable
	}

	if len(s.proxies[country]) == 0 {
		return Proxy{}, ErrNoneAvaliable
	}

	for proxy := range s.proxies[country] {
		return proxy, ErrNoneAvaliable
	}

	return Proxy{}, ErrNoneAvaliable
}

// GetRandomFromCountries gets a random proxy from a given set of countries.
// This method is more expensive than GetRandomFromCountry because it has to first flatten the list of proxies
// from specific countries and then pick one at random.
func (s *Set) GetRandomFromCountries(countries ...string) (Proxy, error) {
	s.m.Lock()
	defer s.m.Unlock()

	proxies := []Proxy{}
	for _, country := range countries {
		if s.proxies[country] == nil {
			continue
		}

		if len(s.proxies[country]) == 0 {
			continue
		}

		for proxy := range s.proxies[country] {
			proxies = append(proxies, proxy)
		}
	}

	if len(proxies) == 0 {
		return Proxy{}, ErrNoneAvaliable
	}

	return proxies[rand.Intn(len(proxies))], nil
}

// Clone returns a copy of itself.
func (s *Set) Clone() *Set {
	set := &Set{
		proxies: make(map[string]map[Proxy]bool),
	}

	set.proxies = s.proxies

	return set
}

// NewSet returns an initialised proxy set.
func NewSet() *Set {
	return &Set{
		proxies: make(map[string]map[Proxy]bool),
	}
}
