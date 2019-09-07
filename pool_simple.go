package prox

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ollybritton/prox/providers"
)

// Pool represents a collection/store of proxies.
type Pool interface {
	Load() error
	Filter() error

	SizeAll() int
	SizeUnused() int

	New() (Proxy, error)
	Random() (Proxy, error)
}

// SimplePool is an implementation of a pool without much added functionality.
// It is a simple wrapper for a provider.
type SimplePool struct {
	provider providers.Provider
	timeout  time.Duration

	All    *providers.Set
	Unused *providers.Set
}

// SizeAll finds the amount of proxies that are currently loaded, used or unused.
func (pool *SimplePool) SizeAll() int {
	return pool.All.Length()
}

// SizeUnused finds the amount of proxies that are currently unused.
func (pool *SimplePool) SizeUnused() int {
	return pool.Unused.Length()
}

// SetTimeout sets a timeout for the provider. By default, it is set to 15s by NewSimplePool.
func (pool *SimplePool) SetTimeout(timeout time.Duration) {
	pool.timeout = timeout
}

// Load fetches the proxies from it's internal provider and stores them.
func (pool *SimplePool) Load() error {
	collector := providers.NewSet()
	ps, err := pool.provider(collector, pool.timeout)
	if err != nil {
		return err
	}

	for _, p := range ps {
		pool.All.Add(p)

		if !pool.Unused.In(p) {
			pool.Unused.Add(p)
		}
	}

	return nil
}

// Random fetches a random proxy. It doesn't care if the proxy has been used already.
// It still marks a proxy as used.
func (pool *SimplePool) Random() (Proxy, error) {
	length := pool.All.Length()

	if length == 0 {
		return Proxy{}, fmt.Errorf("prox (%p): cannot select random proxy, no proxies in pool", pool)
	}

	rawProxy := pool.All.All()[rand.Intn(length)]
	pool.Unused.Remove(rawProxy)

	return *CastProxy(rawProxy), nil
}

// New fetches a new, unused proxy. It returns an error if there are no unused proxies left.
func (pool *SimplePool) New() (Proxy, error) {
	length := pool.Unused.Length()

	if length == 0 {
		return Proxy{}, fmt.Errorf("prox (%p): no unused proxies left in pool", pool)
	}

	rawProxy := pool.Unused.All()[rand.Intn(length)]
	pool.Unused.Remove(rawProxy)

	return *CastProxy(rawProxy), nil
}

// Filter applies the filter to the proxies inside the pool.
func (pool *SimplePool) Filter(filters ...Filter) {
	all := ApplyFilters(pool.All.All(), filters)
	unused := ApplyFilters(pool.Unused.All(), filters)

	pool.All = providers.NewSet()
	pool.Unused = providers.NewSet()

	for _, p := range all {
		pool.All.Add(p)
	}

	for _, p := range unused {
		pool.Unused.Add(p)
	}
}

// NewSimplePool returns a new a new SimplePool struct.
func NewSimplePool(providerNames ...string) *SimplePool {
	for _, providerName := range providerNames {
		panicValidProvider(providerName)
	}

	provider := MultiProvider(providerNames...)

	return &SimplePool{
		provider: provider,
		timeout:  time.Second * 15,

		All:    providers.NewSet(),
		Unused: providers.NewSet(),
	}
}
