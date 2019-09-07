# Providers

This subpackage is what actually finds and downloads the proxies from online proxy providers. This is much lower-level than the encompassing library, and has no implementation of filters or making sure a proxy hasn't been used before.

## Usage

A provider is a function with the following signature:

```go
func(proxies *providers.Set, timeout time.Duration) ([]providers.Proxy, error)
```
