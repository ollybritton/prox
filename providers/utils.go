package providers

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/pariz/gountries"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger
var countryInfo = &countryDB{initialised: false}

// countryDB wraps a geoip2.Reader to make looking up country information easier.
type countryDB struct {
	db          *geoip2.Reader
	query       *gountries.Query
	initialised bool
}

func (cdb *countryDB) FindCountryByIP(ip string) (string, error) {
	if !cdb.initialised {
		panic("providers: cannot access country db, exiting")
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", fmt.Errorf("providers: malformed ip address '%v'", ip)
	}

	record, err := cdb.db.Country(parsedIP)
	if err != nil {
		return "", errors.Wrap(err, "providers: error finding ip in database")
	}

	return record.Country.IsoCode, nil
}

func (cdb *countryDB) FindCountryByName(name string) (string, error) {
	if !cdb.initialised {
		panic("providers: cannot access country db, exiting")
	}

	// edge cases, sometimes a provider provides a string like "Viet Nam" which
	// can't be found, so they are manually added here.
	edge := map[string]string{
		"Korea (South)":               "KR",
		"Great Britain (UK)":          "GB",
		"Viet Nam":                    "VN",
		"New Zealand (Aotearoa)":      "NZ",
		"Croatia (Hrvatska)":          "HR",
		"Cote D'Ivoire (Ivory Coast)": "CI",
		"Congo":                       "CD",
		"European Union":              "EU", // Not a country, but still
	}

	lookup, err := cdb.query.FindCountryByName(name)
	if err != nil {
		if edge[name] != "" {
			return edge[name], nil
		}

		return "", errors.Wrap(err, "providers: error looking up country name")
	}

	return lookup.Codes.Alpha2, nil
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out. Code provided courtesy of stack overflow.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})

	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

// InitLog initialises the logger with options specified.
func InitLog(l *logrus.Logger) {
	logger = l
	log.SetOutput(logger.Writer())
}

// init will initialise the logger when the package is imported and
// open the maxmind database to find country info
func init() {
	l := logrus.New()
	l.SetLevel(logrus.ErrorLevel)

	InitLog(l)

	bytes, err := Asset("data/geo.mmdb")
	if err != nil {
		l.Error(errors.Wrap(err, "providers: cannot access embedded geoip database"))
		return
	}

	internal, err := geoip2.FromBytes(bytes)
	if err != nil {
		l.Error(errors.Wrap(err, "providers: cannot access embedded geoip database"))
		return
	}

	countryInfo = &countryDB{db: internal, query: gountries.New(), initialised: true}

}
