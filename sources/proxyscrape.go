package sources

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

type proxyScrape struct {
	client *http.Client
}

func proxyScrapeURL(options Options) (string, error) {
	var u string

	timeout := options.Timeout.Milliseconds()

	switch strings.ToLower(options.Type) {
	case "http":
		if options.Anonymity == "" {
			options.Anonymity = "all"
		}

		base := "https://api.proxyscrape.com/?request=getproxies&proxytype=%s&timeout=%s&country=%s&ssl=%s&anonymity=%s"
		u = fmt.Sprintf(base, "http", timeout, options.Country, "no", options.Anonymity)

	case "https":
		if options.Anonymity == "" {
			options.Anonymity = "all"
		}

		base := "https://api.proxyscrape.com/?request=getproxies&proxytype=%s&timeout=%s&country=%s&ssl=%s&anonymity=%s"
		u = fmt.Sprintf(base, "http", timeout, options.Country, "yes", options.Anonymity)

	case "socks4":
		if options.Anonymity != "" {
			return "", NewOptionsErr(options, "anonymity has been specifed for socks4 proxy which isn't avaliable with ProxyScrape")
		}

		base := "https://api.proxyscrape.com/?request=getproxies&proxytype=%s&timeout=%s&country=%s"
		u = fmt.Sprintf(base, "socks4", timeout, options.Country)

	case "socks5":
		if options.Anonymity != "" {
			return "", NewOptionsErr(options, "anonymity has been specifed for socks5 proxy which isn't avaliable with ProxyScrape")
		}

		base := "https://api.proxyscrape.com/?request=getproxies&proxytype=%s&timeout=%s&country=%s"
		u = fmt.Sprintf(base, "socks5", timeout, options.Country)
	}

	return u, nil
}

func (s proxyScrape) Query(options Options) ([]Proxy, error) {
	scheme := strings.ToLower(options.Type)

	if scheme != "http" && scheme != "https" && scheme != "socks4" && scheme != "socks5" {
		return []Proxy{}, NewOptionsErr(options, "invalid proxy type "+options.Type)
	}

	u, err := proxyScrapeURL(options)
	if err != nil {
		return []Proxy{}, fmt.Errorf("proxyscrape error: %w", err)
	}

	resp, err := s.client.Get(u)
	if err != nil {
		return []Proxy{}, fmt.Errorf("could not reach proxyscrape API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []Proxy{}, fmt.Errorf("invalid proxyscrape API status code: got %d, wanted %d", resp.StatusCode, http.StatusOK)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	bodyString := string(bodyBytes)
	bodyString = strings.Trim(bodyString, "\n")

	rawIPs := strings.Split(bodyString, "\n")

	proxies := []Proxy{}

	for _, ip := range rawIPs {
		// removes whitespace
		ip = strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, ip)

		u, err := url.Parse(scheme + "://" + ip)
		if err != nil {
			return []Proxy{}, fmt.Errorf("malformed proxyscrape ip '%s': %w", scheme+"://"+ip, err)
		}

		country := ""

		if options.Country == "all" {
			country = "unknown"
		} else {
			country = options.Country
		}

		proxies = append(proxies, Proxy{
			URL:       u,
			Country:   country,
			Anonymity: options.Anonymity,
			Timeout:   options.Timeout,
			Owner:     "ProxyScrape",
		})

	}

	return proxies, nil
}

// OptionsProxyScrape returns an options struct for querying ProxyScrape. If certain values are left empty, it fills them
// with sensible defaults.
func OptionsProxyScrape(t string, country string, anonymity string, timeout time.Duration) Options {
	if country == "" {
		country = "all"
	}

	if timeout == time.Duration(0) {
		timeout = time.Millisecond * 10000
	}

	return Options{
		Type:      t,
		Country:   country,
		Anonymity: anonymity,
		Timeout:   timeout,
	}
}
