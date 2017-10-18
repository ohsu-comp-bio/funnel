package beanstalk

import awsutil "github.com/ohsu-comp-bio/funnel/cmd/aws/util"

type beanstalkConfig struct {
	ApplicationName    string
	EnvironmentName    string
	CNAMEPrefix        string
	SolutionStackName  string
	InstanceType       string
	IamInstanceProfile beanstalkInstanceRole
}

type beanstalkInstanceRole struct {
	Name     string
	Policies struct {
		AssumeRole awsutil.AssumeRolePolicy
		Batch      awsutil.Policy
		DynamoDB   awsutil.Policy
	}
}

func defaultConfig() beanstalkConfig {
	c := beanstalkConfig{
		ApplicationName:   "funnel",
		CNAMEPrefix:       "funnel",
		EnvironmentName:   "funnel",
		SolutionStackName: "64bit Amazon Linux 2017.03 v2.7.4 running Docker 17.03.2-ce",
		InstanceType:      "t2.micro",
	}

	c.IamInstanceProfile.Name = "FunnelEBSInstanceRole"
	c.IamInstanceProfile.Policies.AssumeRole = awsutil.AssumeRolePolicy{
		Version: "2012-10-17",
		Statement: []awsutil.RoleStatement{
			{
				Effect:    "Allow",
				Principal: map[string]string{"Service": "ec2.amazonaws.com"},
				Action:    "sts:AssumeRole",
			},
		},
	}
	c.IamInstanceProfile.Policies.Batch = awsutil.Policy{
		Version: "2012-10-17",
		Statement: []awsutil.Statement{
			{
				Effect: "Allow",
				Action: []string{
					"batch:CancelJob",
					"batch:RegisterJobDefinition",
					"batch:SubmitJob",
				},
				Resource: "*",
			},
		},
	}
	c.IamInstanceProfile.Policies.DynamoDB = awsutil.Policy{
		Version: "2012-10-17",
		Statement: []awsutil.Statement{
			{
				Effect: "Allow",
				Action: []string{
					"dynamodb:CreateTable",
					"dynamodb:GetItem",
					"dynamodb:PutItem",
					"dynamodb:UpdateItem",
					"dynamodb:Query",
				},
				Resource: "*",
			},
		},
	}

	return c
}
