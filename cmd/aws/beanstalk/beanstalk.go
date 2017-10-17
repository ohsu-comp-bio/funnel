package beanstalk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	ebs "github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
	awsutil "github.com/ohsu-comp-bio/funnel/cmd/aws/util"
	"github.com/ohsu-comp-bio/funnel/util"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type svc struct {
	cli     *ebs.ElasticBeanstalk
	sess    *session.Session
	version string
	conf    beanstalkConfig
}

func newSvc(conf beanstalkConfig, version string, region string) (*svc, error) {
	awsConf := util.NewAWSConfigWithCreds("", "")
	awsConf.WithRegion(region)
	sess, err := session.NewSession(awsConf)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating aws session: %v", err)
	}
	return &svc{ebs.New(sess), sess, version, conf}, nil
}

func (s *svc) UploadSourceBundle(ctx context.Context, image string, confPath string, url string) error {
	s3Cli := s3.New(s.sess)

	bundle := strings.TrimPrefix(url, "s3://")
	parts := strings.SplitAfterN(bundle, "/", 2)
	bucket := strings.TrimSuffix(parts[0], "/")
	key := parts[1]

	_, err := s3Cli.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return &awsutil.ErrResourceExists{}
	}

	zipPath, err := ioutil.TempFile("", "funnel-elasticbeanstalk-sourcebundle")
	if err != nil {
		return err
	}
	err = zipPath.Close()
	if err != nil {
		return err
	}
	defer os.Remove(zipPath.Name())

	err = createBundle(zipPath.Name(), image, confPath)
	if err != nil {
		return err
	}

	return uploadBundle(ctx, zipPath.Name(), url)
}

func (s *svc) CreateApplication(ctx context.Context, bundleURL string) (*ebs.ApplicationVersionDescription, error) {
	resp, err := s.cli.DescribeApplicationVersionsWithContext(ctx, &ebs.DescribeApplicationVersionsInput{
		ApplicationName: aws.String(s.conf.ApplicationName),
		VersionLabels:   []*string{aws.String(s.version)},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.ApplicationVersions) > 0 {
		return resp.ApplicationVersions[0], &awsutil.ErrResourceExists{}
	}

	bundle := strings.TrimPrefix(bundleURL, "s3://")
	parts := strings.SplitAfterN(bundle, "/", 2)
	bucket := strings.TrimSuffix(parts[0], "/")
	key := parts[1]

	av, err := s.cli.CreateApplicationVersionWithContext(ctx, &ebs.CreateApplicationVersionInput{
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

func (s *svc) applicationIsActive(ctx context.Context) error {
	ticker := time.NewTicker(time.Millisecond * 500).C
	timeout := time.After(time.Second * 60)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceld")
		case <-timeout:
			return fmt.Errorf("error timeout")
		case <-ticker:
			resp, _ := s.cli.DescribeApplicationVersionsWithContext(ctx, &ebs.DescribeApplicationVersionsInput{
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

func (s *svc) CreateInstanceRole(ctx context.Context) (string, error) {
	iamCli := iam.New(s.sess)

	resp, err := iamCli.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(s.conf.IamInstanceProfile.RoleName),
	})
	if err == nil {
		return *resp.Role.Arn, &awsutil.ErrResourceExists{}
	}

	roleb, err := json.Marshal(s.conf.IamInstanceProfile.Policies.AssumeRole)
	if err != nil {
		return "", fmt.Errorf("error creating AssumeRole policy")
	}

	cr, err := iamCli.CreateRoleWithContext(ctx, &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(string(roleb)),
		RoleName:                 aws.String(s.conf.IamInstanceProfile.RoleName),
	})
	if err != nil {
		return "", err
	}
	return *cr.Role.Arn, nil
}

func (s *svc) AttachRolePolicies(ctx context.Context) error {
	iamCli := iam.New(s.sess)

	resp, err := iamCli.ListRolePolicies(&iam.ListRolePoliciesInput{
		RoleName: aws.String(s.conf.IamInstanceProfile.RoleName),
	})
	if err != nil {
		return err
	}
	if len(resp.PolicyNames) > 0 {
		policies := ""
		for _, v := range resp.PolicyNames {
			policies += *v
		}
		if strings.Contains(policies, "FunnelEBSDynamoDB") && strings.Contains(policies, "FunnelEBSBatch") {
			return &awsutil.ErrResourceExists{}
		}
	}

	batchb, err := json.Marshal(s.conf.IamInstanceProfile.Policies.Batch)
	if err != nil {
		return fmt.Errorf("error creating Batch policy")
	}

	_, err = iamCli.PutRolePolicyWithContext(ctx, &iam.PutRolePolicyInput{
		PolicyDocument: aws.String(string(batchb)),
		PolicyName:     aws.String("FunnelEBSBatch"),
		RoleName:       aws.String(s.conf.IamInstanceProfile.RoleName),
	})
	if err != nil {
		return err
	}

	dynamob, err := json.Marshal(s.conf.IamInstanceProfile.Policies.Batch)
	if err != nil {
		return fmt.Errorf("error creating DynamoDB policy")
	}

	_, err = iamCli.PutRolePolicyWithContext(ctx, &iam.PutRolePolicyInput{
		PolicyDocument: aws.String(string(dynamob)),
		PolicyName:     aws.String("FunnelEBSDynamoDB"),
		RoleName:       aws.String(s.conf.IamInstanceProfile.RoleName),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *svc) CreateEnvironment(ctx context.Context) (*ebs.EnvironmentDescription, error) {
	err := s.applicationIsActive(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := s.cli.DescribeEnvironmentsWithContext(ctx, &ebs.DescribeEnvironmentsInput{
		ApplicationName:  aws.String(s.conf.ApplicationName),
		EnvironmentNames: []*string{aws.String(s.conf.EnvironmentName)},
		IncludeDeleted:   aws.Bool(false),
		VersionLabel:     aws.String(s.version),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Environments) > 0 {
		return resp.Environments[0], &awsutil.ErrResourceExists{}
	}

	instanceRole, err := s.CreateInstanceRole(ctx)
	if err != nil {
		_, ok := err.(awsutil.ErrResourceExists)
		if !ok {
			return nil, err
		}
	}
	err = s.AttachRolePolicies(ctx)
	if err != nil {
		_, ok := err.(awsutil.ErrResourceExists)
		if !ok {
			return nil, err
		}
	}

	return s.cli.CreateEnvironmentWithContext(ctx, &ebs.CreateEnvironmentInput{
		ApplicationName:   aws.String(s.conf.ApplicationName),
		VersionLabel:      aws.String(s.version),
		CNAMEPrefix:       aws.String(s.conf.CNAMEPrefix),
		EnvironmentName:   aws.String(s.conf.EnvironmentName),
		SolutionStackName: aws.String(s.conf.SolutionStackName),
		OptionSettings: []*ebs.ConfigurationOptionSetting{
			{
				Namespace:  aws.String("aws:elasticbeanstalk:application"),
				OptionName: aws.String("Application Healthcheck URL"),
				Value:      aws.String("/health.html"),
			},
			{
				Namespace:  aws.String("aws:autoscaling:launchconfiguration"),
				OptionName: aws.String("IamInstanceProfile"),
				Value:      aws.String(instanceRole),
			},
			{
				Namespace:  aws.String("aws:autoscaling:launchconfiguration"),
				OptionName: aws.String("InstanceType"),
				Value:      aws.String(s.conf.InstanceType),
			},
			// {
			// 	Namespace: aws.String("aws:autoscaling:launchconfiguration"),
			// 	OptionName: aws.String("Custom AMI ID"),
			// 	Value: aws.String(""),
			// },
		},
	})
}
