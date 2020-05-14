package prox

import (
	"fmt"

	"github.com/ollybritton/prox/sources"
)

// SourceOption is a coupling of a source and an option. It is used when passing instructions to a pool.
type SourceOption struct {
	Source  sources.Source
	Options sources.Options
}

// Pool is an automatically updating source of proxies.
type Pool struct {
	sourceOptions []SourceOption

	all    *Set
	unused *Set

	cacheAll    *Set
	cacheUnused *Set
}

// Fetch gets fresh proxies using the sourceOptions specified.
func (p *Pool) Fetch() error {
	p.cacheAll = p.all.Clone()
	p.cacheUnused = p.unused.Clone()

	for _, group := range p.sourceOptions {
		fmt.Println(group)

		proxies, err := group.Source.Query(group.Options)
		if err != nil {
			return fmt.Errorf("could not fetch proxies: %w", err)
		}

		// If a proxy is already in the all set but not in the unused, that means it's been used before so there's no need
		// to add it to anything. Else, add it to the all list and the unused list.
		for _, proxy := range proxies {
			proxy := castProxy(proxy)

			if !(p.all.In(proxy) && !p.unused.In(proxy)) {
				p.all.Add(proxy)
				p.unused.Add(proxy)
			}
		}
	}

	return nil
}

// ArrayAll returns all of the proxies being stored as an array of type []Proxy.
func (p *Pool) ArrayAll() []Proxy {
	return p.all.Array()
}

// ArrayUnused returns all of the unused proxies being stored as an array of type []Proxy.
func (p *Pool) ArrayUnused() []Proxy {
	return p.unused.Array()
}

// Random fetches a random proxy that may or may not have been used.
func (p *Pool) Random() (Proxy, error) {
	proxy, err := p.all.GetRandom()
	if err != nil {
		return Proxy{}, err
	}

	p.unused.Remove(proxy)

	return proxy, nil
}

// RandomFromCountry fetches a random proxy from one of the countries specified.
func (p *Pool) RandomFromCountry(countries ...string) (Proxy, error) {
	proxy, err := p.all.GetRandomFromCountries(countries...)
	if err != nil {
		return Proxy{}, err
	}

	p.unused.Remove(proxy)

	return proxy, nil
}

// NewFromCountry fetches a new, unused proxy from one of the countries specified.
func (p *Pool) NewFromCountry(countries ...string) (Proxy, error) {
	proxy, err := p.unused.GetRandomFromCountries(countries...)
	if err != nil {
		return Proxy{}, err
	}

	p.unused.Remove(proxy)

	return proxy, nil
}

// New fetches a new, unused proxy. Depending on options, it will attempt to reload the proxy
// pool if there are no proxies left inside the pool.
func (p *Pool) New() (Proxy, error) {
	proxy, err := p.unused.GetRandom()
	if err != nil {
		return Proxy{}, err
	}

	p.unused.Remove(proxy)

	return proxy, nil
}

// NewPool returns a new pool built from the given source options.
func NewPool(sourceOptions []SourceOption) *Pool {
	return &Pool{
		sourceOptions: sourceOptions,

		all:         NewSet(),
		unused:      NewSet(),
		cacheAll:    NewSet(),
		cacheUnused: NewSet(),
	}
}
