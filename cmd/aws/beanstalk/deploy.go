package beanstalk

import (
	"context"
	"fmt"
	awsutil "github.com/ohsu-comp-bio/funnel/cmd/aws/util"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/version"
	"github.com/spf13/cobra"
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
	f.StringVar(&bucket, "bucket", bucket, "S3 bucket in which to upload and store the Funnel source bundle")
	f.StringVar(&region, "region", region, "AWS region in which to deploy Funnel server")
}

// DeployCmd represent the command responsible for deploying a Funnel server on
// Amazon ElasticBeanstalk.
var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a Funnel server to Elastic Beanstalk.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return deploy(ctx, image, confPath, bucket, region)
	},
}

func deploy(ctx context.Context, image string, confPath string, bucket string, region string) error {
	log := logger.NewLogger("elasticbeanstalk-deploy", logger.DefaultConfig())

	svc, err := newSvc(defaultConfig(), version.Version, region)
	if err != nil {
		return fmt.Errorf("error creating aws session: %v", err)
	}

	s3Zip := fmt.Sprintf("s3://%s/funnel-app-%s.zip", strings.TrimPrefix(bucket, "s3://"), version.Version)
	err = svc.UploadSourceBundle(ctx, image, confPath, s3Zip)
	switch err.(type) {
	case nil:
		log.Info("Uploaded Application SourceBundle", "url", s3Zip)
	case *awsutil.ErrResourceExists:
		log.Info("Application SourceBundle already exists", "url", s3Zip)
	default:
		return fmt.Errorf("failed to upload application source bundle: %v", err)
	}

	ar, err := svc.CreateApplication(ctx, s3Zip)
	switch err.(type) {
	case nil:
		log.Info("Created Application", "description", ar)
	case *awsutil.ErrResourceExists:
		log.Error("Application already exists", "description", ar)
	default:
		return fmt.Errorf("error creating application: %v", err)
	}

	er, err := svc.CreateEnvironment(ctx)
	switch err.(type) {
	case nil:
		log.Info("Created Application Environment", "description", er)
	case *awsutil.ErrResourceExists:
		log.Error("Application Environment already exists", "description", er)
	default:
		return fmt.Errorf("error creating environment: %v", err)
	}

	return nil
}
