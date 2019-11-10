package cmd

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"github.com/ollybritton/prox"
	"github.com/ollybritton/prox/providers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var errInvalidProvider = errors.New("invalid provider")

// CheckStatus checks the status of a single provider.
func CheckStatus(providerName string) (active bool, amount int, err error) {
	provider := prox.Providers[providerName]

	if provider == nil {
		return false, 0, errInvalidProvider
	}

	set := providers.NewSet()
	timeout := 10 * time.Second

	ps, err := provider(set, timeout)
	if err != nil {
		return false, 0, err
	}

	return true, len(ps), nil
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"check", "monitor"},
	Short:   "get the status of the providers used by prox",
	Long: `get the status of the providers used by prox

To see all providers checked, run
  prox status

To see certain providers checked, run
  prox status --providers FreeProxyLists,ProxyScrape`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()

		providers, err := cmd.Flags().GetStringSlice("providers")
		if err != nil {
			logger.Errorf("couldn't get 'providers' flag: %v", err)
		}

		data := [][]string{}
		total := 0

		avaliable := color.New(color.FgGreen, color.Bold).Sprint("Avaliable")
		unavaliable := color.New(color.FgRed, color.Bold).Sprint("Unavaliable")

		shareColor := color.New(color.FgCyan)

		for _, providerName := range providers {
			logger.Infof("Checking provider %v", providerName)

			_, amount, err := CheckStatus(providerName)
			if err == errInvalidProvider {
				logger.Errorf("No provider named %v", providerName)
				continue
			} else if err != nil {
				logger.Errorf("[%v] Unavaliable", providerName)

				data = append(data, []string{
					providerName, unavaliable, "0", "",
				})

				continue
			}

			data = append(data, []string{
				providerName, avaliable, fmt.Sprint(amount), "",
			})

			total += amount
		}

		if len(data) == 0 {
			return
		}

		for _, entry := range data {
			found, _ := strconv.Atoi(entry[2])
			percentage := 0.0

			if found != 0 {
				percentage = math.Round((float64(found)/float64(total))*10000) / 100
			}

			entry[3] = shareColor.Sprintf("%.2f%%", percentage)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Status", "Proxies Found", "Share"})
		table.SetFooter([]string{"", "TOTAL", fmt.Sprint(total), ""})

		table.SetBorder(false)
		table.SetColumnAlignment([]int{
			tablewriter.ALIGN_DEFAULT,
			tablewriter.ALIGN_DEFAULT,
			tablewriter.ALIGN_LEFT,
			tablewriter.ALIGN_LEFT,
		})

		table.AppendBulk(data)

		fmt.Print("\n\n")
		table.Render()

	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	defaultProviders := []string{}
	for providerName := range prox.Providers {
		if providerName != "Static" {
			defaultProviders = append(defaultProviders, providerName)
		}
	}

	statusCmd.Flags().StringSliceP("providers", "p", defaultProviders, "Provider(s) to check for.")
}
