package providers

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// DummyProvider provides a fixed, small provider of proxies. It is used mainly for testing.
func DummyProvider(proxies *Set, timeout time.Duration) ([]Proxy, error) {
	var ps []Proxy

	staticProxies := []string{
		"http://154.72.199.38:32954 UG",
		"http://117.191.11.103:8080 CN",
		"http://186.103.175.158:3128 CL",
		"http://170.254.150.165:80 BR",
		"http://190.149.165.150:50806 GT",
		"http://112.12.91.67:8888 CN",
		"http://112.12.91.213:8888 CN",
		"http://150.109.55.190:83 US",
		"http://112.12.91.211:8888 CN",
		"http://177.99.11.207:8080 BR",
		"http://110.164.58.180:8080 TH",
		"socks5://178.62.193.19:1080 NL",
		"socks5://82.196.11.105:1080 NL",
		"socks5://94.130.73.24:37188 DE",
		"socks5://94.130.73.24:36975 DE",
		"socks5://94.130.73.24:36966 DE",
		"socks5://94.130.73.24:37374 DE",
		"socks5://54.38.195.161:47404 FR",
		"socks5://94.130.73.30:64918 DE",
	}

	for _, p := range staticProxies {
		columns := strings.Split(p, " ")

		rawurl := columns[0]
		country := columns[1]

		u, err := url.Parse(rawurl)
		if err != nil {
			panic("invalid proxy in dummy provider")
		}

		proxy := Proxy{
			URL:      u,
			Provider: "DummyProvider",
			Country:  country,
		}

		proxies.Add(proxy)
		ps = append(ps, proxy)
	}

	return ps, nil
}

// DummyProviderEmpty provides an example of a provider that is not working, for the purpose of testing.
// It does not return an error.
func DummyProviderEmpty(proxies *Set, timeout time.Duration) ([]Proxy, error) {
	return []Proxy{}, nil
}

// DummyProviderError provides an example of a provider that is not working, for the purpose of testing.
// Unlike DummyProviderNotWoring, this does return a 'no proxies could be gathered' error.
func DummyProviderError(proxies *Set, timeout time.Duration) ([]Proxy, error) {
	return []Proxy{}, fmt.Errorf("providers (DummyProviderError): no proxies could be gathered")
}
