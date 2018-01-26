package cmd

import (
	"fmt"
	"os"

	"github.com/ids/clairctl/clair"
	"github.com/ids/clairctl/config"
	"github.com/ids/clairctl/docker"
	"github.com/ids/clairctl/server"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push IMAGE",
	Short: "Push Docker image to Clair",
	Long:  `Upload a Docker image to Clair for further analysis`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("clairctl: \"push\" requires a minimum of 1 argument\n")
			os.Exit(1)
		}

		if config.IsLocal {
			startLocalServer()
		}
		config.ImageName = args[0]
		image, manifest, err := docker.RetrieveManifest(config.ImageName, true)
		if err != nil {
			fmt.Println(errInternalError)
			log.Fatalf("retrieving manifest for %q: %v", config.ImageName, err)
		}

		if err := clair.Push(image, manifest); err != nil {
			if err != nil {
				fmt.Println(errInternalError)
				log.Fatalf("pushing image %q: %v", image.String(), err)
			}
		}
		fmt.Printf("%v has been pushed to Clair\n", image.String())
	},
}

func startLocalServer() {
	sURL, err := config.LocalServerIP()
	if err != nil {
		fmt.Println(errInternalError)
		log.Fatalf("retrieving internal server IP: %v", err)
	}
	log.Debugf("starting local server with sURL: %v", sURL)
	err = server.Serve(sURL)
	if err != nil {
		fmt.Println(errInternalError)
		log.Fatalf("starting local server: %v", err)
	}
}

func init() {
	RootCmd.AddCommand(pushCmd)
	pushCmd.Flags().BoolVarP(&config.IsLocal, "local", "l", true, "Use local images")
}
