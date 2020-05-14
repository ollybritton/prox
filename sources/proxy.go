package sources

import (
	"net/url"
	"time"
)

// Proxy represents a proxy.
type Proxy struct {
	// URL holds the URL for the proxy. This also contains the scheme, so whether the proxy is HTTP, HTTPS, SOCKS4 or SOCKS5
	URL *url.URL

	// Country is a string containing information about the proxy's origin. It is either an Alpha2 country code or
	// "unknown" if it is not known. Proxies can be labeled "unknown" in two ways. If they are found as part of a query
	// to a source where no country information is specified, then the country of origin cannot be known. They can also
	// be labeled as unknown if the program doesn't understand the country name that a proxy was returned with.
	Country string

	// Anonymity represents how anonymous this proxy is. Some proxy sources make distinctions between proxies that are
	// either "elite", "anonymous" or "transparent". If this information is avaliable, then the anonymity is set to "unknown".
	Anonymity string

	// Timeout represents how fast the proxy is. Some proxy sources let you choose a timeout when fetching proxies.
	Timeout time.Duration

	// Owner is the source that the proxy came from, as a string.
	Owner string
}
