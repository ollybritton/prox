package prox_test

// Some of these tests will only work if internet is available and the providers being used
// are accesible. The status of the providers can be checked with the `prox status` command.

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ollybritton/prox"
	"github.com/ollybritton/prox/providers"
	"github.com/stretchr/testify/assert"
)

var (
	TestProviders  = []string{"FreeProxyLists", "ProxyScrape"}
	DummyProviders = []string{"DummyProvider", "DummyProviderEmpty", "DummyProviderError"}
)

func init() {
	// initialise logging, useful for debugging
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)

	prox.InitLog(logger)

	// add the dummy providers to the provider map
	prox.Providers["DummyProvider"] = providers.DummyProvider
	prox.Providers["DummyProviderEmpty"] = providers.DummyProviderEmpty
	prox.Providers["DummyProviderError"] = providers.DummyProviderError
}

// TestComplexPoolCreation tests that the function NewComplexPool works.
func TestComplexPoolCreation(t *testing.T) {
	pool := prox.NewComplexPool()

	assert.Equal(t, pool.SizeAll(), 0, "unloaded pool should have zero proxies")
	assert.Equal(t, pool.CacheAvailable, false, "cache shouldn't be available for an unloaded pool")
	assert.Equal(t, pool.Config.FallbackToBackupProviders, true)
	assert.Equal(t, pool.Config.FallbackToCached, false)
	assert.Equal(t, pool.Config.ReloadWhenEmpty, false)

	pool = prox.NewComplexPool(
		prox.OptionFallbackToBackupProviders(true),
		prox.OptionFallbackToCached(true),
		prox.OptionReloadWhenEmpty(true),
	)

	assert.Equal(t, pool.Config.FallbackToBackupProviders, true)
	assert.Equal(t, pool.Config.FallbackToCached, true)
	assert.Equal(t, pool.Config.ReloadWhenEmpty, true)
}

// TestComplexPoolProviders tests that loading a pool with a set of providers will cause proxies to be stored.
func TestComplexPoolProviders(t *testing.T) {
	pool := prox.NewComplexPool(prox.UseProviders(TestProviders...))

	t.Logf("Loading 'FreeProxyLists' and 'ProxyScrape' proxies")
	err := pool.Load()

	assert.Nil(t, err, "loading proxies should not produce eerror")
	assert.NotEqual(t, pool.SizeAll(), 0, "amount of proxies loaded should not be zero")

	t.Logf("Found %d proxies in total", pool.SizeAll())
}

// TestComplexPoolInvalidProviders tests that loading a pool with an invalid provider panics.
func TestComplexPoolInvalidProviders(t *testing.T) {
	defer func() {
		err := recover()
		assert.NotNil(t, err, "loading an invalid provider should cause a panic")
	}()

	prox.NewComplexPool(
		prox.UseProviders("HGUIExampleBadProvideraAOIJD"),
	)

}

// TestComplexPoolEmptyProvider tests that using a provider which returns no proxies causes an error.
func TestComplexPoolEmptyProvider(t *testing.T) {
	pool := prox.NewComplexPool(
		prox.UseProvider("DummyProviderEmpty"),
	)

	err := pool.Load()
	assert.NotNil(t, err, "using a provider which returns no proxies should cause an error")
}

// TestComplexPoolErrorProvider tests that using a provider which returns an error causes an error to be
// returned by the pool itself.
func TestComplexPoolErrorProvider(t *testing.T) {
	pool := prox.NewComplexPool(
		prox.UseProvider("DummyProviderError"),
	)

	err := pool.Load()
	assert.NotNil(t, err, "using a provider which returns an error should cause an error")
}

// TestComplexPoolFallbackProviders tests that if the initial provider fails then the backup providers are used.
func TestComplexPoolFallbackProviders(t *testing.T) {
	pool := prox.NewComplexPool(
		prox.UseProvider("DummyProviderEmpty"),
		prox.UseFallbackProvider("DummyProvider"),
	)

	err := pool.Load()
	assert.Nil(t, err, "error should not occur when fallback provider exists")
}

// TestComplexPoolCache tests that a pool will use the cached proxies if the normal providers do not work.
func TestComplexPoolCache(t *testing.T) {
	initialPool := prox.NewComplexPool(
		prox.UseProvider("DummyProvider"),
	)

	err := initialPool.Load()
	assert.Nil(t, err)

	pool := prox.NewComplexPool(
		prox.OptionFallbackToCached(true),
		prox.UseProvider("DummyProviderEmpty"),
	)

	// Load the cache from the initial pool into the new pool.
	// This simulates a provider that stops working.
	pool.CacheAvailable = true
	pool.CacheAll = initialPool.All
	pool.CacheUnused = initialPool.Unused

	err = pool.Load()
	assert.Nil(t, err, "no error should occur if cache is available to the pool")
}

// TestComplexPoolAutomaticReload tests that the pool will be automatically reloaded if the proxy list is empty when
// .New() or .Random() is called, when the ReloadWhenEmpty option is set.
func TestComplexPoolAutomaticReload(t *testing.T) {
	pool := prox.NewComplexPool(
		prox.OptionReloadWhenEmpty(true),
		prox.UseProvider("DummyProvider"),
	)

	_, err := pool.New()
	assert.Nil(t, err, "no error should occur when the pool is automatically reloaded when empty")
	assert.NotEqual(t, pool.SizeAll(), 0, "should not have zero proxies once automatically reloaded")

	pool = prox.NewComplexPool(
		prox.OptionReloadWhenEmpty(true),
		prox.UseProvider("DummyProvider"),
	)

	_, err = pool.Random()
	assert.Nil(t, err, "no error should occur when the pool is automatically reloaded when empty")
	assert.NotEqual(t, pool.SizeAll(), 0, "should not have zero proxies once automatically reloaded")
}

// TestComplexPoolNew tests the .New method of the pool.
func TestComplexPoolNew(t *testing.T) {
	pool := prox.NewComplexPool(
		prox.UseProvider("DummyProvider"),
	)

	err := pool.Load()
	assert.Nil(t, err)

	p, err := pool.New()
	assert.Nil(t, err, "no error should occur when getting new proxy from filled pool")
	assert.NotZero(t, p, "new proxy should not be zero-valued")
	assert.Equal(t, pool.SizeAll()-1, pool.SizeUnused(), "using one proxy should reduce the size of the unused proxy set by 1")

	pool.Unused = providers.NewSet()

	_, err = pool.New()
	assert.NotNil(t, err, "error should occur when fetching a proxy where no proxies are left")
}

// TestComplexPoolRandom tests the .Random method of the pool.
func TestComplexPoolRandom(t *testing.T) {
	pool := prox.NewComplexPool(
		prox.UseProvider("DummyProvider"),
	)

	err := pool.Load()
	assert.Nil(t, err)

	p, err := pool.Random()
	assert.Nil(t, err, "no error should occur when getting random proxy from filled pool")
	assert.NotZero(t, p, "random proxy should not be zero-valued")
	assert.Equal(t, pool.SizeAll()-1, pool.SizeUnused(), "using one proxy should reduce the size of the unused proxy set by 1")

	pool.Unused = providers.NewSet()

	p, err = pool.Random()
	assert.Nil(t, err, "getting random proxy should be independent from the size of the unused proxy")
	assert.NotZero(t, p, "random proxy should not be zero-valued")
}

// TestComplexPoolEquivalentFilters tests that equivalent filters yield equivalent results.
func TestComplexPoolEquivalentFilters(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	pool1 := prox.NewComplexPool(prox.UseProvider("FreeProxyLists"))
	pool2 := prox.NewComplexPool(prox.UseProvider("FreeProxyLists"))

	t.Log("Loading proxy pool...")
	pool1.Load()
	t.Logf("Proxies found: %d", pool1.SizeAll())

	t.Log("Copying proxies from 1st pool to 2nd pool")
	for p := range pool1.All.All() {
		pool2.All.Add(p)
	}

	assert.Equal(t, pool1.SizeAll(), pool2.SizeAll())

	t.Log("Filtering first pool by country and type simultaneously...")
	pool1.Filter(
		prox.FilterProxyTypes("HTTP"),
		prox.FilterDisallowCountries([]string{"BR"}),
	)

	t.Log("Filtering 2nd proxy pool by type and then country")
	pool2.Filter(prox.FilterProxyTypes("HTTP"))
	pool2.Filter(prox.FilterDisallowCountries([]string{"BR"}))

	t.Logf("1st Pool Size: %d", pool1.SizeAll())
	t.Logf("2nd Pool Size: %d", pool2.SizeAll())

	assert.Equal(t, pool1.SizeAll(), pool2.SizeAll())
}

// TestComplexPoolIsStatic tests that after a set of proxies have been loaded, the amount of
// proxies won't increase in the background. This is because it would mess up things like
// filtering.
func TestComplexPoolIsStatic(t *testing.T) {
	pool := prox.NewComplexPool(prox.UseProvider("FreeProxyLists"))

	t.Log("Loading proxy pool...")

	err := pool.Load()
	assert.Nil(t, err)

	initialSize := pool.SizeAll()
	t.Logf("initial size is %d", initialSize)

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		newSize := pool.SizeAll()

		t.Logf("size after %d seconds: %d", i+1, newSize)
		assert.Equal(t, initialSize, newSize)
	}
}
