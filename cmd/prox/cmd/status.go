package cmd

import (
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"github.com/ollybritton/prox"
	"github.com/ollybritton/prox/tools"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "get the status of the providers used by prox",
	Long:  `get the status of the providers used by prox`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()

		data := [][]string{}
		total := 0

		avaliable := color.New(color.FgGreen, color.Bold).Sprint("Avaliable")
		unavaliable := color.New(color.FgRed, color.Bold).Sprint("Unavaliable")

		shareColor := color.New(color.FgCyan)

		for providerName := range prox.Providers {
			logger.Infof("Checking provider %v", providerName)

			_, amount, err := tools.CheckStatus(providerName)
			if err != nil {
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
