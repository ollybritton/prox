package prox

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/imroc/req"
	"github.com/ollybritton/prox/providers"
	"github.com/pkg/errors"
	"golang.org/x/net/proxy"

	// Needed to augment net/proxy to support socks4
	_ "github.com/Bogdan-D/go-socks4"
)

// Proxy holds information about a proxy. It is like providers.Proxy,
// but it contains more methods.
type Proxy struct {
	URL      *url.URL
	Provider string
	Country  string

	used bool

	client    *http.Client
	hasClient bool
}

// PrettyPrint prints some information in a nicely-formatted way.
func (p *Proxy) PrettyPrint() {
	urlString := p.URL.String()

	fmt.Println(
		fmt.Sprintf("(%v) %v [%v]", p.Country, urlString, p.Provider),
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

	p.client = client
	p.hasClient = true

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
// timeout, and will mark a proxy as unavaliable if it doesn't respond within that time.
func (p *Proxy) CheckSpeed(timeout time.Duration) bool {
	r := req.New()

	client, err := p.Client()
	if err != nil {
		return false
	}

	prev := client.Timeout
	client.Timeout = timeout
	defer func() { client.Timeout = prev }()

	r.SetClient(client)

	_, err = r.Get("https://example.org")
	if err != nil {
		return false
	}

	return true
}

// CheckConnection checks that a connection to a proxy can be formed. It is agnostic to timeouts.
func (p *Proxy) CheckConnection() bool {
	r := req.New()

	client, err := p.Client()
	if err != nil {
		return false
	}

	prev := client.Timeout
	client.Timeout = 10 * time.Second
	defer func() { client.Timeout = prev }()

	r.SetClient(client)

	_, err = r.Get("https://example.org")
	if err != nil {

		// This is bad
		if strings.Contains(err.Error(), "Client.Timeout") {
			return true
		}

		return false
	}

	return true
}

// CastProxy will convert a providers.Proxy type into prox.Proxy type.
func CastProxy(p providers.Proxy) *Proxy {
	return &Proxy{
		URL:      p.URL,
		Provider: p.Provider,
		Country:  p.Country,
	}
}

// NewProxy will create a new proxy type.
func NewProxy(rawip string, provider string, country string) (Proxy, error) {
	u, err := url.Parse(rawip)
	if err != nil {
		return Proxy{}, errors.Wrap(err, "prox: cannot parse ip to URL")
	}

	return Proxy{
		URL:      u,
		Provider: provider,
		Country:  country,
	}, nil
}
