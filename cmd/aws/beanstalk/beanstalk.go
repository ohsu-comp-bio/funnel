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
	"time"
)

type svc struct {
	ctx     context.Context
	cli     *ebs.ElasticBeanstalk
	version string
	conf    beanstalkConfig
}

func newSvc(ctx context.Context, conf beanstalkConfig, version string, region string) (*svc, error) {
	awsConf := util.NewAWSConfigWithCreds("", "")
	awsConf.WithRegion(region)
	sess, err := session.NewSession(awsConf)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating aws session: %v", err)
	}
	return &svc{ctx, ebs.New(sess), version, conf}, nil
}

func (s *svc) uploadBundle(src string, dest string) error {
	s3, err := storage.NewS3Backend(config.S3Storage{})
	if err != nil {
		return err
	}
	_, err = s3.Put(s.ctx, dest, src, tes.FileType_FILE)
	return err
}

func (s *svc) createApplication(bundleURL string) (*ebs.ApplicationVersionDescription, error) {
	resp, err := s.cli.DescribeApplicationVersionsWithContext(s.ctx, &ebs.DescribeApplicationVersionsInput{
		ApplicationName: aws.String(s.conf.ApplicationName),
		VersionLabels:   []*string{aws.String(s.version)},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.ApplicationVersions) > 0 {
		return resp.ApplicationVersions[0], &errResourceExists{}
	}

	bundle := strings.TrimPrefix(bundleURL, "s3://")
	parts := strings.SplitAfterN(bundle, "/", 2)
	bucket := strings.TrimSuffix(parts[0], "/")
	key := parts[1]

	av, err := s.cli.CreateApplicationVersionWithContext(s.ctx, &ebs.CreateApplicationVersionInput{
		ApplicationName:       aws.String(s.conf.ApplicationName),
		AutoCreateApplication: aws.Bool(true),
		Process:               aws.Bool(true),
		SourceBundle: &ebs.S3Location{
			S3Bucket: aws.String(bucket),
			S3Key:    aws.String(key),
		},
		VersionLabel: aws.String(s.version),
	})
	return av.ApplicationVersion, err
}

func (s *svc) applicationIsActive() error {
	ticker := time.NewTicker(time.Millisecond * 500).C
	timeout := time.After(time.Second * 60)

	for {
		select {
		case <-s.ctx.Done():
			return fmt.Errorf("context canceld")
		case <-timeout:
			return fmt.Errorf("error timeout")
		case <-ticker:
			resp, _ := s.cli.DescribeApplicationVersionsWithContext(s.ctx, &ebs.DescribeApplicationVersionsInput{
				ApplicationName: aws.String(s.conf.ApplicationName),
				VersionLabels:   []*string{aws.String(s.version)},
			})
			if len(resp.ApplicationVersions) > 0 {
				switch *resp.ApplicationVersions[0].Status {
				case "PROCESSED":
					return nil
				case "FAILED":
					return fmt.Errorf("application processing failed")
				}
			}
		}
	}
}

func (s *svc) createEnvironment() (*ebs.EnvironmentDescription, error) {
	err := s.applicationIsActive()
	if err != nil {
		return nil, err
	}

	resp, err := s.cli.DescribeEnvironmentsWithContext(s.ctx, &ebs.DescribeEnvironmentsInput{
		ApplicationName:  aws.String(s.conf.ApplicationName),
		EnvironmentNames: []*string{aws.String(s.conf.EnvironmentName)},
		IncludeDeleted:   aws.Bool(false),
		VersionLabel:     aws.String(s.version),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Environments) > 0 {
		return resp.Environments[0], &errResourceExists{}
	}

	return s.cli.CreateEnvironmentWithContext(s.ctx, &ebs.CreateEnvironmentInput{
		ApplicationName:   aws.String(s.conf.ApplicationName),
		CNAMEPrefix:       aws.String(s.conf.CNAMEPrefix),
		EnvironmentName:   aws.String(s.conf.EnvironmentName),
		VersionLabel:      aws.String(s.version),
		SolutionStackName: aws.String(s.conf.SolutionStackName),
	})
}
