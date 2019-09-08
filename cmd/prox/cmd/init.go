package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/ollybritton/prox/tools"
	"github.com/sirupsen/logrus"

	"github.com/mitchellh/go-homedir"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "download the MaxMind GeoLite2 database for IP country filtering",
	Long:  `initialise the tool by setting up the `,

	Run: func(cmd *cobra.Command, args []string) {
		dbfolder, err := cmd.Flags().GetString("path")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		homedir, err := homedir.Dir()
		if err != nil {
			fmt.Println("could not locate home directory:")
			fmt.Println(err)
			os.Exit(1)
		}

		if dbfolder == "" {
			folder := path.Join(homedir, ".config", "prox")
			err = os.MkdirAll(folder, os.ModePerm)

			if err != nil {
				fmt.Println("could not make db folder", folder)
				fmt.Println(err)
				os.Exit(1)
			}

			dbfolder = folder
		}

		err = tools.SetupDatabase(logrus.New(), homedir, dbfolder)
		if err != nil {
			fmt.Println("\nDB installation failed.")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("path", "", "Where to download the database to. Defaults to $HOME/.config/prox/geo.mmdb")
}
