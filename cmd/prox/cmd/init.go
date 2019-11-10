package cmd

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/mitchellh/go-homedir"

	"github.com/spf13/cobra"
)

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
// Code courtest of https://golangcode.com/download-a-file-from-a-url/.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// SetupDatabase downloads the MaxMind GeoLite2 database and moves it to the specified directory.
// It uses a separate logger as it isn't involved in proxy finding.
func SetupDatabase(logger *logrus.Logger, homedir string, dbfolder string) error {
	dbpath := path.Join(dbfolder, "geo.tar.gz")

	logger.Infof("downloading DB archive to %v folder", dbpath)
	err := DownloadFile(dbpath, "https://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.tar.gz")

	if err != nil {
		logger.Errorf("error occurred during download: %v", err)
		return errors.Wrap(err, "error occurred during download")
	}

	logger.Info("unarchiving geo.tar.gz")
	err = archiver.Unarchive(dbpath, path.Join(dbfolder, "geo"))

	if err != nil {
		logger.Errorf("error occurred while unarchiving: %v", err)
		return errors.Wrap(err, "error occurred while unarchiving")
	}

	logger.Info("locating geo.mmdb file in unzipped download")

	files, err := ioutil.ReadDir(path.Join(dbfolder, "geo"))
	if err != nil {
		logger.Errorf("error occurred while reading files inside unzipped directory: %v", err)
		return errors.Wrap(err, "error occurred while reading files inside unzipped directory")
	}

	found := false

	for _, f := range files {
		if f.IsDir() {
			os.Rename(path.Join(dbfolder, "geo", f.Name(), "GeoLite2-Country.mmdb"), path.Join(dbfolder, "geo.mmdb"))

			found = true
			break
		}
	}

	if !found {
		logger.Errorf("couldn't find GeoLite2-Country.mmdb inside unzipped archive")
		return errors.Wrap(err, "couldn't find GeoLite2-Country.mmdb inside unzipped archive")
	}

	var shellConfigFile string

	switch os.Getenv("SHELL") {
	case "/bin/zsh":
		shellConfigFile = path.Join(homedir, ".zshrc")
	case "/bin/bash":
		shellConfigFile = path.Join(homedir, ".bashrc")
	default:
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Unable to detect your shell config file automatically.")
		fmt.Print("[prox] Please enter your shell config file (such as .bashrc): ")
		text, _ := reader.ReadString('\n')

		shellConfigFile = path.Join(homedir, text)
	}

	logger.Infof("setting $PROX_GEODB environment variable inside %v", shellConfigFile)

	f, err := os.OpenFile(shellConfigFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Errorf("cannot access shell config file %v", shellConfigFile)
		return errors.Wrap(err, "cannot access shell config file "+shellConfigFile)
	}
	defer f.Close()

	code := fmt.Sprintf("\nexport PROX_GEODB=%v\n", path.Join(dbfolder, "geo.mmdb"))
	if _, err := f.WriteString(code); err != nil {
		logger.Errorf("cannot access shell config file %v", shellConfigFile)
		return errors.Wrap(err, "cannot access shell config file "+shellConfigFile)
	}

	logger.Infof("cleaning up %v", dbfolder)

	err = os.Remove(path.Join(dbfolder, "geo.tar.gz"))
	if err != nil {
		logger.Errorf("error removing archive: %v", err)
		return errors.Wrap(err, "error removing archive")
	}

	err = os.RemoveAll(path.Join(dbfolder, "geo"))
	if err != nil {
		logger.Errorf("error removing unzipped archive: %v", err)
		return errors.Wrap(err, "error removing unzipped archive")
	}

	logger.Info("Database successfully downloaded.")
	return nil
}

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

		err = SetupDatabase(logrus.New(), homedir, dbfolder)
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
