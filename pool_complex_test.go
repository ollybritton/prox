package prox_test

import (
	"testing"

	"github.com/ollybritton/prox"
	"github.com/stretchr/testify/assert"
)

// TestComplexPoolProviders tests that a complex pool loaded with a set of providers works.
func TestComplexPoolProviders(t *testing.T) {
	pool := prox.NewComplexPool(
		prox.UseProviders("FreeProxyLists", "ProxyScrape"),
	)

	t.Logf("Loading 'FreeProxyLists' and 'ProxyScrape' proxies")
	err := pool.Load()

	assert.Nil(t, err)
	assert.NotEqual(t, pool.SizeAll(), 0)

	t.Logf("Found %d proxies in total", pool.SizeAll())

}
