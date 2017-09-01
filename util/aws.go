package util

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// NewAWSSession returns and new github.com/aws/aws-sdk-go/aws/session.Session
// instance. If both the key and secret are empty AWS credentials will
// be read from the environment.
func NewAWSSession(key string, secret string, region string) (*session.Session, error) {
	awsConf := aws.NewConfig().WithRegion(region)

	if key != "" && secret != "" {
		creds := credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     key,
			SecretAccessKey: secret,
		})
		awsConf.WithCredentials(creds)
	}

	return session.NewSession(awsConf)
}
