package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ohsu-comp-bio/funnel/config"
)

// NewAWSSession returns a new session.Session instance.
func NewAWSSession(conf config.AWSConfig) (*session.Session, error) {
	awsConf := aws.NewConfig()

	if conf.Endpoint != "" {
		awsConf.WithEndpoint(conf.Endpoint)
	}

	if conf.Region != "" {
		awsConf.WithRegion(conf.Region)
	}

	if conf.MaxRetries > 0 {
		awsConf.WithMaxRetries(conf.MaxRetries)
	}

	if conf.Key != "" && conf.Secret != "" {
		creds := credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     conf.Key,
			SecretAccessKey: conf.Secret,
		})
		awsConf.WithCredentials(creds)
	}

	return session.NewSession(awsConf)
}
