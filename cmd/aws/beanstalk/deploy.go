package beanstalk

import (
	// "fmt"
	// "github.com/ohsu-comp-bio/funnel/logger"
	"github.com/spf13/cobra"
	"path"
)

var image = "docker.io/ohsu-comp-bio/funnel:latest"
var region = "us-west-2"
var confPath = ""

func init() {
	f := DeployCmd.Flags()
	f.StringVar(&image, "image", image, "Funnel docker image to deploy")
	f.StringVar(&confPath, "config", confPath, "Config file")
	f.StringVar(&region, "region", region, "AWS region in which to deploy Funnel server")
}

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a Funnel server to Elastic Beanstalk.",
	RunE: func(cmd *cobra.Command, args []string) error {

		zip = path.Join(".", "funnel_elasticbeanstalk.zip")
		if val := os.Getenv("TMPDIR"); val != "" {
			zip = path.Join(val, "funnel_elasticbeanstalk.zip")
		}

		return createBundle(zip, image, confPath)
	},
}
