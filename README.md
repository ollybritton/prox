# Prox
Prox is a simple Go package for locating open proxy servers. It works by congregating proxies from several different providers on the internet and allows access to them using a simple API. It is the successor to my previous package, [proxyfinder](https://github.com/ollybritton/proxyfinder).

[![GoDoc](https://godoc.org/github.com/ollybritton/prox?status.svg)](https://godoc.org/github.com/ollybritton/prox)
[![Go Report Card](https://goreportcard.com/badge/github.com/ollybritton/prox)](https://goreportcard.com/report/github.com/ollybritton/prox)

- [Prox](#prox)
  - [Setup](#setup)
  - [Command-Line Tool](#command-line-tool)
  - [Usage](#usage)
    - [Key Terms](#key-terms)
    - [High Level (Pools)](#high-level-pools)
      - [Simple Pool](#simple-pool)
      - [Complex Pool](#complex-pool)
    - [Low Level (Providers & Sets)](#low-level-providers--sets)
      - [The `providers.Proxy` type](#the-providersproxy-type)
      - [The `providers.Set` type](#the-providersset-type)
      - [The `providers.Provider` type](#the-providersprovider-type)
  - [Bugs](#bugs)
  - [Legal](#legal)

## Setup
Assuming you have a proper go install, you can just run
```bash
$ go get -u github.com/ollybritton/prox
```

This will install the package, as well as the `prox` [command line tool]((#command-line-tool)). The MaxMind GeoIP database is embedded into the binary, which means no additional setup is required. Although this does increase the size (20M for the binary!), I would argue that this is a neccessary evil as it reduces setup and besides, the database would need to stored somewhere anyway.



## Command-Line Tool
There are a few more useful features of the command line tool:

```bash
$ prox status # Check status of providers
$ prox find # Print proxies to the terminal
```

For help about a specific command, just do

```bash
$ prox help <command name>
```

## Usage
The library provides a high level interface (`Pools`) and a lower level interface `Providers` to the proxy providers. For most uses, the higher-level `Pool` implementation is better.

### Key Terms
A **Provider** is just a source of proxies. It can be a website or a static list stored somewhere on the machine. The following providers are available.

| Name             | Source                                                          |
| ---------------- | --------------------------------------------------------------- |
| `FreeProxyLists` | [http://freeproxylists.com/](http://freeproxylists.com/)        |
| `ProxyScrape`    | [https://proxyscrape.com/](https://proxyscrape.com/)`           |
| `Static`         | Stored in `providers/data/proxies`, accessed using `go-bindata` |

### High Level (Pools)
Pools are simply a collection of [providers](#providers) combined together that can keep track of proxies that have been used and those that haven't. There are two types of pools, `SimplePools` and `ComplexPools`.

#### Simple Pool
To create a new `SimplePool`, use the `NewSimplePool` function supplied with a name or a list of names of providers.
```go
pool := prox.NewSimplePool(prox.FreeProxyLists)
```

You then need to load the pool:
```go
pool := prox.NewSimplePool(prox.FreeProxyLists)
if err := pool.Load(); err != nil {
    panic(err)
}
```

By default, loading the proxies will take a maximum of about 15 seconds. Most of the time, it is much faster than this. The following methods are then available:
```go
proxy, err := pool.New() // Fetch a new, unused proxy. Will error if there are no unused proxies left.
proxy, err := pool.Random() // Fetch a random proxy, used or unused. It will still be marked as used so you won't be able to access this proxy with pool.New()

pool.SetTimeout(10 * time.Second) // Set the maximum timeout of the proxy list.

pool.SizeAll() // Get the amount of proxies in the pool.
pool.SizeUnused() // Get the amount of unused proxies in the pool.

pool.Filter(
    prox.FilterAllowCountries([]string{"GB", "US"}) // Only allow the specified countries in the pool
    prox.FilterDisallowCountries([]string{"GB", "US"}) // Allow anything but the specified countries.
    prox.FilterProxyConnection() // Only allow proxies that can be connected to. If they take longer than 10 seconds to connect to, they are PRESUMED TO BE WORKING.
    prox.FilterProxySpeed(5 * time.Second) // Only allow proxies that can be connected to in the given timeframe. Presumed to not be working if it takes longer than the timeout.
    prox.FilterProxyTypes("HTTP", "SOCKS4", "SOCKS5") // Only allow proxies of those types in the pool.
)
```

Note that a filter only applies to the proxies that are currently loaded. If you call `.Load()` again, proxies which don't fit the filters given are still allowed into the pool.

The proxies themselves (the ones returned after a call to `.New()` or `.Random()`) have to following methods:

```go
proxy, err := pool.New()
if err != nil {
    panic(err)
}

canConnect := proxy.CheckConnection() // Checks a proxy can be connected to. Again, it is PRESUMED TO BE WORKING if it cannot connect in 10 seconds. This isn't ideal.
canConnectSpeed := proxy.CheckSpeed(5 * time.Second) // Checks a proxy can be connected to in a given timeframe. 
httpClient := proxy.Client() // Gets the proxy as a *http.Client.
proxy.PrettyPrint() // Prints a proxy's info.
```

#### Complex Pool
`ComplexPools` are like `SimplePools`, but contain more options for things such as automatically refreshing the pool if it is empty and having fallback providers for if the primary ones do not work.

A `ComplexPool` can be created with the `NewComplexPool` function or `NewPool` for short.

```go
pool := prox.NewPool(
    prox.UseProvider(prox.FreeProxyLists), // Use a provider, or...
    prox.UseProviders(prox.FreeProxyLists, prox.ProxyScrape) // a list of providers

    prox.UseFallbackProviders(prox.Static), // Provider to "fall back" on if the primary providers do not work or return an error.
    prox.OptionFallbackToBackupProviders(true), // Toggle this option. By default it is true.

    prox.OptionFallbackToCached(true), // Keep a backup of the previously loaded proxies. If the providers can't be accessed, use the cached list of proxies instead. Defaults to false.

    prox.OptionReloadWhenEmpty(true), // If there are no proxies left in the pool when .New() or .Random() are called, load the proxies again. Defaults to false.

    prox.OptionAddFilters(
        // filters are identical to the ones used in SimplePool
        // These filters will be called everytime the pool is loaded, unlike SimplePool
    )
)
```

The following methods are then available for the `ComplexPool`

```go
err := pool.Load() // Load the proxies.


proxy, err := pool.New() // Fetch a new, unused proxy. Will error if there are no unused 
proxy, err := pool.Random() // Fetch a random proxy, used or unused. It will still be marked as used so you won't be able to access this proxy with pool.New()

pool.Option([options go here]) // Set another option on the pool
pool.Filter([filter name]) // Apply a filter to the proxies in the pool. These are not permanent.

pool.SetTimeout(10 * time.Second) // Set a timeout for fetching the proxies.

pool.SizeAll() // Size of all proxies.
pool.SizeUnused() // Size of unused proxies.

err := pool.ApplyCache() // Use the previously available cache. It will error if there is not a cache available.
```

### Low Level (Providers & Sets)
A lower level interface to the proxy providers is also available, available through the `providers/` package. In reality, the `Pool` implementation wraps the providers package to provide the additional functionality.

#### The `providers.Proxy` type
The `providers.Proxy` type is basically identical to the `prox.Proxy` type. The only reason the `prox.Proxy` type exists is to provide additional functionality to the lower level implementation (like the `.Client()` method) and prevent circular imports. 

It has the following fields:

```go
type Proxy struct {
    URL      *url.URL `json:"url"`
    Provider string   `json:"providers"`
    Country  string   `json:"country"`

    Used bool
}
```

#### The `providers.Set` type
The `providers.Set` type is a concurrency-safe set implementation which can store proxies. It means that proxies can be stored asynchronously and not cause race conditions.

```go
set := providers.NewSet()

set.Add(p providers.Proxy) // Add a proxy to the set
set.All() // Get all the proxies in the set as a slice
set.In(p providers.Proxy) // Check if a proxy is the set
set.Length() // Get the length of the set
set.Remove(p providers.Proxy) // Remove a proxy from the set
```

#### The `providers.Provider` type
This is the definition of a provider. It has the following type signature:

```go
type Provider func(*providers.Set, timeout time.Duration) ([]providers.Proxy, error)
```

Simply put, it is a function which takes a [set](#the-providersset-type) and a timeout and returns a list of proxies and an error if one occurs.

For example, the implementation of `FreeProxyLists` is:

```go
func FreeProxyLists(proxies *providers.Set, timeout time.Duration) ([]providers.Proxy, error) {
    // do stuff like proxies.Add()
    return proxies.All(), nil
}
```

The reason it is implemented this way is so that the list of proxies can be accessed as the function runs.

```go
set := providers.NewSet()
go FreeProxyLists(set, 10 * time.Seconds)

for {
    println(set.Length())
    time.Sleep(1 * time.Second)
}

// Output:
// 0
// 0
// 0
// 231
// 873
// 2563
// ...
```


## Bugs
* HTTPS proxies
  
  Creating HTTPS clients doesn't work, always either 
  * `net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)`, or
  * `proxyconnect tcp: tls: first record does not look like a TLS handshake`

  The consequence of this is that all HTTPS proxies are marked as unaccessible when filtering. To remove all HTTPS proxies so that you don't run into this problem
  
  ```go
  pool.Filter(prox.FilterProxyTypes("HTTP", "SOCKS4", "SOCKS5"))
  ```
  
## Legal
> This product includes GeoLite2 data created by MaxMind, available from [https://www.maxmind.com](https://www.maxmind.com).
