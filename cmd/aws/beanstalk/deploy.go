package beanstalk

import (
	"context"
	"fmt"
	// "github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/version"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
)

var image = "docker.io/ohsucompbio/funnel:latest"
var region = "us-west-2"
var bucket = "strucka-dev"
var confPath = ""

func init() {
	f := DeployCmd.Flags()
	f.StringVar(&image, "image", image, "Funnel docker image to deploy")
	f.StringVar(&confPath, "config", confPath, "Config file")
	f.StringVar(&bucket, "bucket", bucket, "S3 bucket in which to store Funnel server logs")
	f.StringVar(&region, "region", region, "AWS region in which to deploy Funnel server")
}

// DeployCmd represent the command responsible for deploying a Funnel server on
// Amazon ElasticBeanstalk.
var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a Funnel server to Elastic Beanstalk.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return deploy(image, confPath, bucket, region)
	},
}

func deploy(image string, confPath string, bucket string, region string) error {
	zip := path.Join(".", "funnel_elasticbeanstalk.zip")
	if val := os.Getenv("TMPDIR"); val != "" {
		zip = path.Join(val, "funnel_elasticbeanstalk.zip")
	}

	bucket = strings.TrimPrefix(bucket, "s3://")
	s3Zip := fmt.Sprintf("s3://%s/funnel-app-%s.zip", bucket, version.Version)
	err := createBundle(zip, image, confPath)
	if err != nil {
		return fmt.Errorf("error creating platform bundle: %v", err)
	}

	svc, err := newSvc(context.Background(), defaultConfig(), version.Version, region)
	if err != nil {
		return fmt.Errorf("error creating aws session: %v", err)
	}
	err = svc.uploadBundle(zip, s3Zip)
	if err != nil {
		return fmt.Errorf("error uploading platform bundle: %v", err)
	}

	ar, err := svc.createApplication(s3Zip)
	if err != nil {
		fmt.Println(fmt.Errorf("error creating application: %v", err))
	}
	fmt.Println(ar)
	er, err := svc.createEnvironment()
	if err != nil {
		return fmt.Errorf("error creating environment: %v", err)
	}
	fmt.Println(er)

	return err
}
