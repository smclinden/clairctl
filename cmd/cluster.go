package cmd

import (
	"github.com/ids/clairctl/clair"
	"github.com/spf13/cobra"
	"github.com/ids/clairctl/config"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Scan and analyze all Docker images in cluster",
	Long:  `Scan and analyze all Docker images in cluster, against Ubuntu, Red hat and Debian vulnerabilities databases`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		images := clair.ClusterScan();
		config.ClusterImages = images
	},
}

func init() {
	RootCmd.AddCommand(clusterCmd)
}