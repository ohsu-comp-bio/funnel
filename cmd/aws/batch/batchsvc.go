package batch

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/ohsu-comp-bio/funnel/util"
	"strings"
)

type errResourceExists struct{}

func (e errResourceExists) Error() string {
	return "resource exists"
}

func newBatchSvc(conf Config) (*batchsvc, error) {
	awsConf := util.NewAWSConfigWithCreds("", "")
	awsConf.WithRegion(conf.Region)
	sess, err := session.NewSession(awsConf)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating aws session: %v", err)
	}
	return &batchsvc{
		sess: sess,
		conf: conf,
	}, nil
}

type batchsvc struct {
	sess *session.Session
	conf Config
}

func (b *batchsvc) CreateComputeEnvironment() (*batch.CreateComputeEnvironmentOutput, error) {
	batchCli := batch.New(b.sess)
	ec2Cli := ec2.New(b.sess)
	iamCli := iam.New(b.sess)

	resp, err := batchCli.DescribeComputeEnvironments(&batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: []*string{aws.String(b.conf.ComputeEnv.Name)},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.ComputeEnvironments) > 0 {
		return &batch.CreateComputeEnvironmentOutput{
			ComputeEnvironmentArn:  resp.ComputeEnvironments[0].ComputeEnvironmentArn,
			ComputeEnvironmentName: resp.ComputeEnvironments[0].ComputeEnvironmentName,
		}, &errResourceExists{}
	}

	securityGroupIds := []string{}
	if len(b.conf.ComputeEnv.SecurityGroupIds) > 0 {
		securityGroupIds = b.conf.ComputeEnv.SecurityGroupIds
	} else {
		sgres, err := ec2Cli.DescribeSecurityGroups(nil)
		if err != nil {
			return nil, err
		}
		for _, s := range sgres.SecurityGroups {
			securityGroupIds = append(securityGroupIds, *s.GroupId)
		}

	}

	subnets := []string{}
	if len(b.conf.ComputeEnv.Subnets) > 0 {
		subnets = b.conf.ComputeEnv.Subnets
	} else {
		snres, err := ec2Cli.DescribeSubnets(nil)
		if err != nil {
			return nil, err
		}
		for _, s := range snres.Subnets {
			subnets = append(subnets, *s.SubnetId)
		}
	}

	var serviceRole string
	iamres, err := iamCli.GetRole(&iam.GetRoleInput{RoleName: aws.String("AWSBatchServiceRole")})
	if err == nil {
		serviceRole = *iamres.Role.Arn
	} else {
		bsrPolicy := AssumeRolePolicy{
			Version: "2012-10-17",
			Statement: []RoleStatement{
				{
					Effect:    "Allow",
					Principal: map[string]string{"Service": "batch.amazonaws.com"},
					Action:    "sts:AssumeRole",
				},
			},
		}
		bsrBinary, err := json.Marshal(bsrPolicy)
		if err != nil {
			return nil, fmt.Errorf("error marshaling assume role policy for AWSBatchServiceRole: %v", err)
		}
		cr, err := iamCli.CreateRole(&iam.CreateRoleInput{
			AssumeRolePolicyDocument: aws.String(string(bsrBinary)),
			RoleName:                 aws.String("AWSBatchServiceRole"),
		})
		if err != nil {
			return nil, fmt.Errorf("error creating AWSBatchServiceRole: %v", err)
		}
		_, err = iamCli.AttachRolePolicy(&iam.AttachRolePolicyInput{
			PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSBatchServiceRole"),
			RoleName:  aws.String("AWSBatchServiceRole"),
		})
		if err != nil {
			return nil, fmt.Errorf("error attaching policies to AWSBatchServiceRole: %v", err)
		}
		serviceRole = *cr.Role.Arn
	}

	var instanceRole string
	iamres, err = iamCli.GetRole(&iam.GetRoleInput{RoleName: aws.String("ecsInstanceRole")})
	if err == nil {
		instanceRole = *iamres.Role.Arn
	} else {
		irPolicy := AssumeRolePolicy{
			Version: "2012-10-17",
			Statement: []RoleStatement{
				{
					Effect:    "Allow",
					Principal: map[string]string{"Service": "ec2.amazonaws.com"},
					Action:    "sts:AssumeRole",
				},
			},
		}
		irBinary, err := json.Marshal(irPolicy)
		if err != nil {
			return nil, fmt.Errorf("error marshaling assume role policy for ecsInstanceRole: %v", err)
		}
		cr, err := iamCli.CreateRole(&iam.CreateRoleInput{
			AssumeRolePolicyDocument: aws.String(string(irBinary)),
			RoleName:                 aws.String("ecsInstanceRole"),
		})
		if err != nil {
			return nil, fmt.Errorf("error creating ecsInstanceRole: %v", err)
		}
		_, err = iamCli.AttachRolePolicy(&iam.AttachRolePolicyInput{
			PolicyArn: aws.String("arn:aws:iam::aws:policy/AmazonEC2ContainerServiceforEC2Role"),
			RoleName:  aws.String("ecsInstanceRole"),
		})
		if err != nil {
			return nil, fmt.Errorf("error attaching policies to ecsInstanceRole: %v", err)
		}
		instanceRole = *cr.Role.Arn
	}

	input := &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(b.conf.ComputeEnv.Name),
		ComputeResources: &batch.ComputeResource{
			InstanceRole:     aws.String(instanceRole),
			InstanceTypes:    convertStringSlice(b.conf.ComputeEnv.InstanceTypes),
			MaxvCpus:         aws.Int64(b.conf.ComputeEnv.MaxVCPUs),
			MinvCpus:         aws.Int64(b.conf.ComputeEnv.MinVCPUs),
			SecurityGroupIds: convertStringSlice(securityGroupIds),
			Subnets:          convertStringSlice(subnets),
			Tags:             convertStringMap(b.conf.ComputeEnv.Tags),
			Type:             aws.String("EC2"),
		},
		ServiceRole: aws.String(serviceRole),
		State:       aws.String("ENABLED"),
		Type:        aws.String("MANAGED"),
	}

	return batchCli.CreateComputeEnvironment(input)
}

func (b *batchsvc) CreateJobQueue() (*batch.CreateJobQueueOutput, error) {
	batchCli := batch.New(b.sess)

	resp, err := batchCli.DescribeJobQueues(&batch.DescribeJobQueuesInput{
		JobQueues: []*string{aws.String(conf.JobQueue.Name)},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.JobQueues) > 0 {
		return &batch.CreateJobQueueOutput{
			JobQueueArn:  resp.JobQueues[0].JobQueueArn,
			JobQueueName: resp.JobQueues[0].JobQueueName,
		}, &errResourceExists{}
	}

	var envs []*batch.ComputeEnvironmentOrder
	for i, c := range b.conf.JobQueue.ComputeEnvs {
		envs = append(envs, &batch.ComputeEnvironmentOrder{
			ComputeEnvironment: aws.String(c),
			Order:              aws.Int64(int64(i)),
		})
	}

	input := &batch.CreateJobQueueInput{
		ComputeEnvironmentOrder: envs,
		JobQueueName:            aws.String(b.conf.JobQueue.Name),
		Priority:                aws.Int64(b.conf.JobQueue.Priority),
		State:                   aws.String("ENABLED"),
	}

	return batchCli.CreateJobQueue(input)
}

func (b *batchsvc) CreateJobRole() (*iam.CreateRoleOutput, error) {
	iamCli := iam.New(b.sess)

	resp, err := iamCli.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(b.conf.JobRole.RoleName),
	})
	if err == nil {
		return &iam.CreateRoleOutput{
			Role: resp.Role,
		}, &errResourceExists{}
	}

	roleb, err := json.Marshal(b.conf.JobRole.Policies.AssumeRole)
	if err != nil {
		return nil, fmt.Errorf("error creating assume role policy")
	}

	cr, err := iamCli.CreateRole(&iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(string(roleb)),
		RoleName:                 aws.String(b.conf.JobRole.RoleName),
	})
	if err != nil {
		return nil, err
	}

	return cr, nil
}

func (b *batchsvc) AttachRolePolicies() error {
	iamCli := iam.New(b.sess)

	resp, err := iamCli.ListRolePolicies(&iam.ListRolePoliciesInput{
		RoleName: aws.String(b.conf.JobRole.RoleName),
	})
	if err != nil {
		return err
	}
	if len(resp.PolicyNames) > 0 {
		policies := ""
		for _, v := range resp.PolicyNames {
			policies += *v
		}
		if strings.Contains(policies, b.conf.JobRole.DynamoDBPolicyName) && strings.Contains(policies, b.conf.JobRole.S3PolicyName) {
			return &errResourceExists{}
		}
	}

	s3b, err := json.Marshal(b.conf.JobRole.Policies.S3)
	if err != nil {
		return fmt.Errorf("error creating s3 policy")
	}

	dynb, err := json.Marshal(b.conf.JobRole.Policies.DynamoDB)
	if err != nil {
		return fmt.Errorf("error creating dynamodb policy")
	}

	_, err = iamCli.PutRolePolicy(&iam.PutRolePolicyInput{
		RoleName:       aws.String(b.conf.JobRole.RoleName),
		PolicyDocument: aws.String(string(s3b)),
		PolicyName:     aws.String(b.conf.JobRole.S3PolicyName),
	})
	if err != nil {
		return err
	}

	_, err = iamCli.PutRolePolicy(&iam.PutRolePolicyInput{
		RoleName:       aws.String(b.conf.JobRole.RoleName),
		PolicyDocument: aws.String(string(dynb)),
		PolicyName:     aws.String(b.conf.JobRole.DynamoDBPolicyName),
	})
	if err != nil {
		return err
	}

	return nil
}

func convertStringSlice(s []string) []*string {
	var ret []*string
	for _, t := range s {
		ret = append(ret, aws.String(t))
	}
	return ret
}

func convertStringMap(s map[string]string) map[string]*string {
	m := map[string]*string{}
	for k, v := range s {
		m[k] = aws.String(v)
	}
	return m
}
