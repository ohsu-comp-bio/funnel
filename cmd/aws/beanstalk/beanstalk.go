package beanstalk

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	ebs "github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/util"
	"strings"
)

type svc struct {
	sess *session.Session
	ctx  context.Context
}

func newSvc(ctx context.Context, region string) (*svc, error) {
	awsConf := util.NewAWSConfigWithCreds("", "")
	awsConf.WithRegion(region)
	sess, err := session.NewSession(awsConf)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating aws session: %v", err)
	}
	return &svc{sess, ctx}, nil
}

func (s *svc) uploadBundle(src string, dest string) error {
	s3, err := storage.NewS3Backend(config.S3Storage{})
	if err != nil {
		return err
	}
	_, err = s3.Put(s.ctx, dest, src, tes.FileType_FILE)
	return err
}

func (s *svc) createApplication(bundleUrl string, version string) (*ebs.ApplicationVersionDescriptionMessage, error) {
	cli := ebs.New(s.sess)

	bundle := strings.TrimPrefix(bundleUrl, "s3://")
	parts := strings.SplitAfterN(bundle, "/", 2)
	bucket := strings.TrimSuffix(parts[0], "/")
	key := parts[1]

	return cli.CreateApplicationVersion(&ebs.CreateApplicationVersionInput{
		ApplicationName:       aws.String("funnel-server"),
		AutoCreateApplication: aws.Bool(true),
		Description:           aws.String("Funnel server backended by DynamoDB and Batch"),
		Process:               aws.Bool(true),
		SourceBundle: &ebs.S3Location{
			S3Bucket: aws.String(bucket),
			S3Key:    aws.String(key),
		},
		VersionLabel: aws.String(version),
	})
}
