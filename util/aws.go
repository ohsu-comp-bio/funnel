package util

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// NewAWSConfigWithCreds returns a new aws.Config instance. If both the key
// and secret are empty AWS credentials will be read from the environment.
func NewAWSConfigWithCreds(key string, secret string) *aws.Config {
	awsConf := aws.NewConfig()

	if key != "" && secret != "" {
		creds := credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     key,
			SecretAccessKey: secret,
		})
		awsConf.WithCredentials(creds)
	}

	return awsConf
}
