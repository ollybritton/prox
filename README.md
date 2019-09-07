# Prox
Prox is a simple Go package for locating open proxy servers. It works by congregating proxies from several different providers on the internet and allows access to them using a simple API. It is the successor to my previous package, [proxyfinder](https://github.com/ollybritton/proxyfinder).

- [Prox](#prox)
  - [Setup](#setup)
  - [Usage](#usage)
    - [Providers](#providers)
  - [Bugs](#bugs)

## Setup
This package requires you have a *MaxMind GeoLite2* or *GeoIP2* database installed on your system. This can be achieved by running `prox init` once installed.

## Usage
### Providers
A provider is just a function which returns a list of 
```go
import "github.com/ollybritton/prox"

func main() {
    pool := prox.NewPool()

    pool := prox.NewPool(prox.UseProvider("https://freeproxylists.com"))

    pool := prox.NewPool(
        prox.UseProviders("https://freeproxylists.com", "https://proxyscrape.com", "static"1)
    )
}
```

## Bugs
* HTTPS proxies
  Creating HTTP_S_ clients doesn't work, always either 
  * `net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)`, or
  * `proxyconnect tcp: tls: first record does not look like a TLS handshake`

  The consequence of this is that all HTTPS proxies are marked as unaccessible when filtering.
  