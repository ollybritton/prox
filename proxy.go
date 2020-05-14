package prox

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ollybritton/prox/sources"
	"golang.org/x/net/proxy"

	// Needed to augment net/proxy to support socks4
	_ "github.com/Bogdan-D/go-socks4"
)

// Proxy is a wrapper around a sources.Proxy but provides additional functionality.
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

	client    *http.Client
	hasClient bool
	used      bool
}

// PrettyPrint prints some information in a nicely-formatted way.
func (p *Proxy) PrettyPrint() {
	urlString := p.URL.String()

	fmt.Println(
		fmt.Sprintf("(%v) %v", p.Country, urlString),
	)
}

// Client gets the http.Client associated with the given proxy.
func (p *Proxy) Client() (*http.Client, error) {
	if p.hasClient {
		return p.client, nil
	}

	var client *http.Client
	var err error

	switch p.URL.Scheme {
	case "http":
		client, err = p.AsHTTPClient()
	case "https":
		client, err = p.AsHTTPSClient()
	case "socks4":
		client, err = p.AsSOCKS4Client()
	case "socks5":
		client, err = p.AsSOCKS5Client()
	default:
		panic("prox: unknown proxy type " + p.URL.Scheme)
	}

	if err != nil {
		return &http.Client{}, err
	}

	return client, nil
}

// AsHTTPClient will return the proxy as a http.Client struct.
// It panics if the proxy's type is not HTTP.
func (p *Proxy) AsHTTPClient() (*http.Client, error) {
	if p.URL.Scheme != "http" {
		panic(
			fmt.Errorf("prox: cannot get HTTP client of %v proxy", p.URL.Scheme),
		)
	}

	client := &http.Client{}
	client.Transport = &http.Transport{Proxy: http.ProxyURL(p.URL)}

	return client, nil
}

// AsHTTPSClient will return the proxy as a http.Client struct.
// It panics if the proxy's type is not HTTPS.
func (p *Proxy) AsHTTPSClient() (*http.Client, error) {
	if p.URL.Scheme != "https" {
		panic(
			fmt.Errorf("prox: cannot get HTTPS client of %v proxy", p.URL.Scheme),
		)
	}

	client := &http.Client{}
	client.Transport = &http.Transport{
		Proxy:        http.ProxyURL(p.URL),
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	return client, nil
}

// AsSOCKS4Client will return the proxy as a http.Client struct.
// It panics if the proxy's type is not SOCKS4.
func (p *Proxy) AsSOCKS4Client() (*http.Client, error) {
	if p.URL.Scheme != "socks4" {
		panic(
			fmt.Errorf("prox: cannot get SOCKS4 client of %v proxy", p.URL.Scheme),
		)
	}

	dialer, err := proxy.FromURL(p.URL, proxy.Direct)
	if err != nil {
		return &http.Client{}, fmt.Errorf("prox: cannot connect to socks4 proxy: %v", err)
	}

	transport := &http.Transport{}
	transport.Dial = dialer.Dial

	client := &http.Client{Transport: transport}

	return client, nil
}

// AsSOCKS5Client will return the proxy as a http.Client struct.
// It panics if the proxy's type is not SOCKS5.
func (p *Proxy) AsSOCKS5Client() (*http.Client, error) {
	if p.URL.Scheme != "socks5" {
		panic(
			fmt.Errorf("prox: cannot get SOCKS5 client of %v proxy", p.URL.Scheme),
		)
	}

	dialer, err := proxy.FromURL(p.URL, proxy.Direct)
	if err != nil {
		return &http.Client{}, fmt.Errorf("prox: cannot connect to socks5 proxy: %v", err)
	}

	transport := &http.Transport{}
	transport.Dial = dialer.Dial

	client := &http.Client{Transport: transport}

	return client, nil
}

// CheckSpeed checks that a connection to proxy can be formed. It accepts a
// timeout, and will mark a proxy as unavailable if it doesn't respond within that time.
func (p *Proxy) CheckSpeed(timeout time.Duration) bool {
	client, err := p.Client()
	if err != nil {
		return false
	}

	prev := client.Timeout
	client.Timeout = timeout
	defer func() { client.Timeout = prev }()

	resp, err := client.Get("http://gstatic.com/generate_204")
	if err != nil {
		return false
	}

	if resp.StatusCode != 204 {
		return false
	}

	return true
}

// CheckConnection checks that a connection to a proxy can be formed. It will still mark a proxy as successful even if
// it times out. If you want to filter proxies that timeout, use CheckSpeed(10 * time.Second), which is equivalent.
func (p *Proxy) CheckConnection() bool {
	client, err := p.Client()
	if err != nil {
		return false
	}

	prev := client.Timeout
	client.Timeout = 10 * time.Second
	defer func() { client.Timeout = prev }()

	resp, err := client.Get("http://gstatic.com/generate_204")
	if err != nil {
		if err == http.ErrHandlerTimeout {
			return true
		}

		return false
	}

	if resp.StatusCode != 204 {
		return false
	}

	return true
}

func castProxy(proxy sources.Proxy) Proxy {
	return Proxy{
		URL:       proxy.URL,
		Country:   proxy.Country,
		Anonymity: proxy.Anonymity,
		Owner:     proxy.Owner,
		Timeout:   proxy.Timeout,

		client:    nil,
		hasClient: false,
		used:      false,
	}
}
