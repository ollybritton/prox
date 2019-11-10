package cmd

import (
	"fmt"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/ollybritton/prox"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func pprint(p prox.Proxy) {
	var country string

	if len(p.Country) != 2 {
		country = "??"
	} else {
		country = p.Country
	}

	fmt.Println(
		aurora.Sprintf(
			aurora.White("(%v) %v [%v]"),
			aurora.Green(country).Bold(),
			aurora.BrightWhite(p.URL.String()),
			aurora.Magenta(p.Provider).Italic(),
		),
	)
}

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:     "find",
	Short:   "find and print proxies",
	Long:    `find and print proxies`,
	Aliases: []string{"fetch", "search"},
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()

		providers, err := cmd.Flags().GetStringSlice("providers")
		if err != nil {
			logger.Errorf("couldn't get 'providers' flag: %v", err)
			return
		}

		types, err := cmd.Flags().GetStringSlice("types")
		if err != nil {
			logger.Errorf("couldn't get types flag: %v", err)
			return
		}

		duration, err := cmd.Flags().GetDuration("duration")
		if err != nil {
			logger.Errorf("couldn't get duration flag: %v", err)
			return
		}

		n, err := cmd.Flags().GetInt("number")
		if err != nil {
			logger.Errorf("couldn't get number flag: %v", err)
			return
		}

		plain, err := cmd.Flags().GetBool("plain")
		if err != nil {
			logger.Errorf("couldn't get plain flag: %v", err)
			return
		}

		pool := prox.NewComplexPool(
			prox.UseProviders(providers...),
			prox.OptionReloadWhenEmpty(true),

			prox.OptionAddFilters(
				prox.FilterProxyTypes(types...),
			),
		)

		pool.SetTimeout(duration)

		for i := 0; i < n; i++ {
			p, err := pool.New()
			if err != nil {
				logger.Errorf("error fetching proxies: %v", err)
				return
			}

			if !plain {
				pprint(p)
			} else {
				fmt.Println(p.URL)
			}

		}
	},
}

func init() {
	rootCmd.AddCommand(findCmd)

	defaultProviders := []string{}
	for providerName := range prox.Providers {
		if providerName != "Static" {
			defaultProviders = append(defaultProviders, providerName)
		}
	}

	findCmd.Flags().StringSliceP("providers", "p", defaultProviders, "providers to fetch")
	findCmd.Flags().StringSliceP("types", "t", []string{"HTTP", "SOCKS4", "SOCKS5"}, "proxy types to fetch")

	findCmd.Flags().DurationP("duration", "d", 10*time.Second, "duration to fetch for")
	findCmd.Flags().IntP("number", "n", 100, "number of proxies to return")

	findCmd.Flags().Bool("plain", false, "use a plain output")
}
