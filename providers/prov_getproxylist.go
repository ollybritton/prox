package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type getProxyListResponse struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Country  string `json:"country"`
}

func getProxyListWorker(id int, num int, client *http.Client, results chan Proxy) {
	for i := 0; i < num; i++ {
		resp, err := client.Get("https://api.getproxylist.com/proxy")
		if err != nil {
			logger.Debugf("providers (GetProxyList): cannot request GetProxyList endpoint: %v", err)
			return
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Debugf("providers (GetProxyList): cannot read response body from GetProxyList: %v", err)
			continue
		}

		response := &getProxyListResponse{}

		err = json.Unmarshal(bytes, response)
		if err != nil {
			logger.Debugf("providers (GetProxyList): could not unmarshal api response")
			continue
		}

		proxy, err := newProxy(response.IP, "GetProxyList", response.Country)
		if err != nil {
			logger.Debugf("providers (GetProxyList): cannot create new proxy: %v", err)
			continue
		}

		results <- proxy
	}
}

// GetProxyList returns the proxies that can be found on the site https://api.getproxylist.com/proxy.
func GetProxyList(proxies *Set, timeout time.Duration) ([]Proxy, error) {
	logger.Debug("providers: Fetching proxies from provider GetProxyList")
	client := &http.Client{}

	results := make(chan Proxy, 100)

	for i := 0; i < 100; i++ {
		go getProxyListWorker(i, 50*50, client, results)
	}

	tchan := make(chan struct{})

	go func() {
		for proxy := range results {
			proxies.Add(proxy)
		}

		tchan <- struct{}{}
	}()

	select {
	case <-tchan:
		break
	case <-time.After(timeout):
		break
	}

	ps := proxies.List()
	if len(ps) == 0 {
		return ps, fmt.Errorf("providers (GetProxyList): no proxies could be gathered")
	}

	return ps, nil
}
