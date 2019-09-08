package tools

import (
	"fmt"
	"time"

	"github.com/ollybritton/prox"
	"github.com/ollybritton/prox/providers"
)

// CheckStatus checks the status of a single provider.
func CheckStatus(providerName string) (active bool, amount int, err error) {
	provider := prox.Providers[providerName]

	if provider == nil {
		return false, 0, fmt.Errorf("invalid provider: %v", providerName)
	}

	set := providers.NewSet()
	timeout := 10 * time.Second

	ps, err := provider(set, timeout)
	if err != nil {
		return false, 0, err
	}

	return true, len(ps), nil
}
