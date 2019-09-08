package prox_test

import (
	"testing"
	"time"

	"github.com/ollybritton/prox"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestSimplePoolIsStatic tests that after a set of proxies have been loaded, the amount of
// proxies won't increase in the background. This is because it would mess up things like
// filtering.
func TestSimplePoolIsStatic(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	pool := prox.NewSimplePool("FreeProxyLists")

	t.Log("Loading proxy pool...")
	pool.Load()

	initialSize := pool.SizeAll()
	t.Logf("initial size is %d", initialSize)

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		newSize := pool.SizeAll()

		t.Logf("size after %d seconds: %d", i+1, newSize)
		assert.Equal(t, initialSize, newSize)
	}
}

// TestSimplePoolEquivialentFilters tests that equivialent filters yield equivalent results.
func TestSimplePoolEquivialentFilters(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	pool1 := prox.NewSimplePool("FreeProxyLists")
	pool2 := prox.NewSimplePool("FreeProxyLists")

	t.Log("Loading proxy pool...")
	err := pool1.Load()
	assert.Nil(t, err)

	t.Logf("Proxies found: %d", pool1.SizeAll())

	t.Log("Copying proxies from 1st pool to 2nd pool")
	for _, p := range pool1.All.All() {
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
