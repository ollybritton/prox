package main

import (
	"time"

	prox "github.com/ollybritton/prox2"
	"github.com/ollybritton/prox2/sources"
	"github.com/sirupsen/logrus"
)

func main() {
	americas := []string{"US", "CA", "MX", "AR", "DE"}
	europe := []string{"AR", "DE", "BE", "NL", "ES", "IT", "CH", "AU", "FR"}
	uk := []string{"UK", "IE", "GB"}

	options := []prox.SourceOption{}

	for _, country := range americas {
		options = append(
			options,
			prox.SourceOption{sources.ProxyScrape, sources.OptionsProxyScrape("http", country, "", time.Second*20)},
			prox.SourceOption{sources.ProxyScrape, sources.OptionsProxyScrape("https", country, "", time.Second*20)},
			prox.SourceOption{sources.ProxyScrape, sources.OptionsProxyScrape("socks4", country, "", time.Second*20)},
			prox.SourceOption{sources.ProxyScrape, sources.OptionsProxyScrape("socks5", country, "", time.Second*20)},
		)
	}

	for _, country := range europe {
		options = append(
			options,
			prox.SourceOption{sources.ProxyScrape, sources.OptionsProxyScrape("http", country, "", time.Second*20)},
			prox.SourceOption{sources.ProxyScrape, sources.OptionsProxyScrape("https", country, "", time.Second*20)},
			prox.SourceOption{sources.ProxyScrape, sources.OptionsProxyScrape("socks4", country, "", time.Second*20)},
			prox.SourceOption{sources.ProxyScrape, sources.OptionsProxyScrape("socks5", country, "", time.Second*20)},
		)
	}

	for _, country := range uk {
		options = append(
			options,
			prox.SourceOption{sources.ProxyScrape, sources.OptionsProxyScrape("http", country, "", time.Second*20)},
			prox.SourceOption{sources.ProxyScrape, sources.OptionsProxyScrape("https", country, "", time.Second*20)},
			prox.SourceOption{sources.ProxyScrape, sources.OptionsProxyScrape("socks4", country, "", time.Second*20)},
			prox.SourceOption{sources.ProxyScrape, sources.OptionsProxyScrape("socks5", country, "", time.Second*20)},
		)
	}

	pool := prox.NewPool(options)

	err := pool.Fetch()
	if err != nil {
		logrus.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		p, err := pool.NewFromCountry(uk...)
		if err != nil {
			logrus.Error(err)
		}

		logrus.Printf("%v\n", p)
	}
}
