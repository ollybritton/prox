package prox

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/ollybritton/prox/providers"
)

// ComplexPool is an implementation of a pool with lots of extra settings, including filtering
// and ensuring the pool always has proxies to provide.
type ComplexPool struct {
	providers         []providers.Provider
	fallbackProviders []providers.Provider
	timeout           time.Duration

	filters []Filter

	Config struct {
		FallbackToBackupProviders bool
		FallbackToCached          bool

		ReloadWhenEmpty bool
	}

	All    *providers.Set
	Unused *providers.Set

	CacheAvailable bool
	CacheAll       *providers.Set
	CacheUnused    *providers.Set
}

// SizeAll finds the amount of proxies that are currently loaded, used or unused.
func (pool *ComplexPool) SizeAll() int {
	return pool.All.Length()
}

// SizeUnused finds the amount of proxies that are currently unused.
func (pool *ComplexPool) SizeUnused() int {
	return pool.Unused.Length()
}

// SetTimeout sets a timeout for the provider. By default, it is set to 15s by NewSimplePool.
func (pool *ComplexPool) SetTimeout(timeout time.Duration) {
	logger.Debugf("prox (%p): setting timeout: %v", pool, timeout)
	pool.timeout = timeout
}

// Fetch fetches the proxies from it's internal providers and stores them.
func (pool *ComplexPool) Fetch() error {
	logger.Debugf("prox (%p): attempting to fetch proxies from providers", pool)
	collector := providers.NewSet()
	wg := &sync.WaitGroup{}

	for _, provider := range pool.providers {
		wg.Add(1)
		go func(provider providers.Provider) {
			defer wg.Done()

			ps, err := provider(collector, pool.timeout)
			if err != nil {
				log.Println(err)

			}

			for _, p := range ps {
				pool.All.Add(p)

				if !pool.Unused.In(p) {
					pool.Unused.Add(p)
				}
			}

		}(provider)

	}

	wg.Wait()

	if len(collector.All()) == 0 {
		logger.Errorf("prox (%p): no proxies could be loaded from providers", pool)
		return fmt.Errorf("prox (%p): no proxies could be loaded from providers", pool)
	}

	logger.Debugf("prox (%p): fetched %d proxies", pool, len(collector.All()))
	logger.Debugf("prox (%p): updating cache with new proxies", pool)
	pool.CacheAvailable = true
	pool.CacheAll = pool.All
	pool.CacheUnused = pool.Unused

	return nil
}

// FetchFallback fetches the proxies from it's fallback providers and stores them.
func (pool *ComplexPool) FetchFallback() error {
	logger.Debugf("prox (%p): attempting to fetch proxies from fallback providers", pool)
	collector := providers.NewSet()
	wg := &sync.WaitGroup{}

	for _, provider := range pool.fallbackProviders {
		wg.Add(1)
		go func(provider providers.Provider) {
			defer wg.Done()

			ps, err := provider(collector, pool.timeout)
			if err != nil {
				log.Println(err)

			}

			for _, p := range ps {
				pool.All.Add(p)

				if !pool.Unused.In(p) {
					pool.Unused.Add(p)
				}
			}

		}(provider)

	}

	wg.Wait()

	if len(collector.All()) == 0 {
		logger.Errorf("prox (%p): no proxies could be fetched from fallback providers", pool)
		return fmt.Errorf("prox (%p): no proxies could be loaded from fallback providers", pool)
	}

	logger.Debugf("prox (%p): fetched %d proxies", pool, len(collector.All()))

	return nil
}

// ApplyCache will revert the pool to the previous cache.
func (pool *ComplexPool) ApplyCache() error {
	if !pool.CacheAvailable {
		return fmt.Errorf("prox (%p): no cache to revert back to", pool)
	}

	pool.All = pool.CacheAll
	pool.Unused = pool.CacheUnused

	pool.CacheAvailable = false

	return nil
}

// Load will fetch the proxies like a call to Fetch(), but, depending on options, it will fallback to a proxy
// cache or use the fallback providers.
func (pool *ComplexPool) Load() error {
	logger.Debugf("prox (%p): attempting to load new proxies", pool)

	err := pool.Fetch()
	if err == nil {
		pool.Filter(pool.filters...)
		return nil
	}

	if pool.Config.FallbackToCached {
		logger.Errorf("prox (%p): error occurred while fetching proxies: %v", pool, err)

		err := pool.ApplyCache()
		if err == nil {
			pool.Filter(pool.filters...)
			return nil
		}

		logger.Errorf("prov (%p): could not apply cache", pool)
	}

	if pool.Config.FallbackToBackupProviders {
		logger.Errorf("prox (%p): error occurred while fetching proxies: %v", pool, err)
		logger.Errorf("prox (%p): falling back to fallback providers", pool)

		err = pool.FetchFallback()
		if err == nil {
			pool.Filter(pool.filters...)
			return nil
		}

		logger.Errorf("prox (%p): error occurred while fetching fallback proxies: %v", pool, err)
	}

	return err
}

// Random fetches a random proxy. It doesn't care if the proxy has been used already.
// It still marks a proxy as used.
func (pool *ComplexPool) Random() (Proxy, error) {
	length := pool.SizeAll()

	if length == 0 {
		if !pool.Config.ReloadWhenEmpty {
			return Proxy{}, fmt.Errorf("prox (%p): cannot select random proxy, no proxies in pool", pool)
		}

		err := pool.Load()
		if err != nil {
			return Proxy{}, fmt.Errorf("prox (%p): cannot select random proxy, error occurred while reloading empty pool: %v", pool, err)
		}

		length = pool.SizeAll()
	}

	rawProxy := pool.All.All()[rand.Intn(length)]
	pool.Unused.Remove(rawProxy)

	return *CastProxy(rawProxy), nil
}

// New fetches a new, unused proxy. Depending on options, it will attempt to reload the proxy
// pool if there are no proxies left inside the pool.
func (pool *ComplexPool) New() (Proxy, error) {
	length := pool.SizeUnused()

	if length == 0 {
		if !pool.Config.ReloadWhenEmpty {
			return Proxy{}, fmt.Errorf("prox (%p): cannot select proxy, no unused proxies left in pool", pool)
		}

		err := pool.Load()
		if err != nil {
			return Proxy{}, fmt.Errorf("prox (%p): cannot select unused proxy, error occurred while reloading pool: %v", pool, err)
		}

		length = pool.SizeUnused()
		if length == 0 {
			return Proxy{}, fmt.Errorf("prox (%p): cannot select proxy, no unused proxies even after reload", pool)
		}
	}

	rawProxy := pool.Unused.All()[rand.Intn(length)]
	pool.Unused.Remove(rawProxy)

	return *CastProxy(rawProxy), nil
}

// Filter applies the filter to the proxies inside the pool.
func (pool *ComplexPool) Filter(filters ...Filter) {
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

// NewComplexPool creates a new complex pool from the options given and using defaults if
// options aren't provided.
func NewComplexPool(opts ...Option) *ComplexPool {
	pool := &ComplexPool{
		All:     providers.NewSet(),
		Unused:  providers.NewSet(),
		timeout: 15 * time.Second,
	}

	// Default config options
	pool.Config.FallbackToBackupProviders = true

	logger.Infof("prox: created new complex pool with id %p", pool)

	for _, opt := range opts {
		opt(pool)
	}

	return pool
}

// NewPool creates a new complex pool from the options given and using defaults if options aren't provided.
// It's an alias for NewComplexPool.
func NewPool(opts ...Option) *ComplexPool {
	return NewComplexPool(opts...)
}

// Option is an option that can be provided to configure a complex pool.
// See https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html for more info
type Option func(*ComplexPool) error

// UseProviders will adds providers to the pool. If any of the provider names is invalid, it will
// panic.
func UseProviders(providerNames ...string) Option {
	return func(p *ComplexPool) error {
		for _, providerName := range providerNames {
			provider := Providers[providerName]

			if provider == nil {
				logger.Panicf("prox (%p): invalid provider '%v'", p, providerName)
			}

			p.providers = append(p.providers, provider)
		}

		logger.Debugf("prox (%p): using providers %v", p, providerNames)

		return nil
	}
}

// UseProvider will add a provider to pool. If the provider name is invalid, it will panic.
func UseProvider(providerName string) Option {
	return UseProviders(providerName)
}

// UseFallbackProviders adds providers that will only be used if the other providers do not work.
// If any of the provider names are invalid, it will panic.
func UseFallbackProviders(providerNames ...string) Option {
	return func(p *ComplexPool) error {
		for _, providerName := range providerNames {
			provider := Providers[providerName]

			if provider == nil {
				logger.Panicf("prox (%p): invalid fallback provider '%v'", p, providerName)
			}

			p.fallbackProviders = append(p.fallbackProviders, provider)
		}

		logger.Debugf("prox (%p): using fallback providers %v", p, providerNames)

		return nil
	}
}

// UseFallbackProvider adds a provider that will only be used if the other providers fail.
// If the provider name is invalid, it will panic.
func UseFallbackProvider(providerName string) Option {
	return UseFallbackProviders(providerName)
}

// OptionReloadWhenEmpty sets the option to attempt to load new proxies into the pool if there are no proxies left in
// the pool on a call to .Random() or .New()
func OptionReloadWhenEmpty(setting bool) Option {
	return func(pool *ComplexPool) error {
		pool.Config.ReloadWhenEmpty = setting
		return nil
	}
}

// OptionFallbackToCached sets the option to use cached proxies when there is an error during loading.
func OptionFallbackToCached(setting bool) Option {
	return func(pool *ComplexPool) error {
		pool.Config.FallbackToCached = setting
		return nil
	}
}

// OptionFallbackToBackupProviders sets the option to use the fallback providers if there is an error during loading.
func OptionFallbackToBackupProviders(setting bool) Option {
	return func(pool *ComplexPool) error {
		pool.Config.FallbackToBackupProviders = setting
		return nil
	}
}

// OptionAddFilters adds a list of filters to the pool
func OptionAddFilters(filters ...Filter) Option {
	return func(pool *ComplexPool) error {
		pool.filters = append(pool.filters, filters...)
		return nil
	}
}

// OptionAddFilter adds a single filter to the pool.
func OptionAddFilter(filter Filter) Option {
	return func(pool *ComplexPool) error {
		pool.filters = append(pool.filters, filter)
		return nil
	}
}

// Option sets the pool options specified.
func (pool *ComplexPool) Option(opts ...Option) (err error) {
	for _, opt := range opts {
		err = opt(pool)
	}

	return err
}

func init() {
	// Initialiase random number generator so that the same proxies aren't picked every time.
	rand.Seed(time.Now().UTC().UnixNano())
}
