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
)

func newBatchSvc(conf Config, dryRun bool) (*batchsvc, error) {
	sess, err := util.NewAWSSession(conf.Key, conf.Secret, conf.Region)
	if err != nil {
		return nil, fmt.Errorf("error occurred creating aws session: %v", err)
	}
	return &batchsvc{
		sess:   sess,
		conf:   conf,
		dryRun: dryRun,
	}, nil
}

type batchsvc struct {
	sess   *session.Session
	conf   Config
	dryRun bool
}

func (b *batchsvc) CreateComputeEnvironment() (*batch.CreateComputeEnvironmentOutput, error) {
	batchCli := batch.New(b.sess)
	ec2Cli := ec2.New(b.sess)
	iamCli := iam.New(b.sess)

	sgres, err := ec2Cli.DescribeSecurityGroups(nil)
	if err != nil {
		return nil, err
	}

	securityGroupIds := []string{}
	for _, s := range sgres.SecurityGroups {
		securityGroupIds = append(securityGroupIds, *s.GroupId)
	}

	snres, err := ec2Cli.DescribeSubnets(nil)
	if err != nil {
		return nil, err
	}

	subnets := []string{}
	for _, s := range snres.Subnets {
		subnets = append(subnets, *s.SubnetId)
	}

	iamres, err := iamCli.GetRole(&iam.GetRoleInput{RoleName: aws.String("AWSBatchServiceRole")})
	if err != nil {
		return nil, err
	}

	serviceRole := *iamres.Role.Arn
	input := &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(b.conf.ComputeEnv.Name),
		ComputeResources: &batch.ComputeResource{
			InstanceRole:     aws.String(b.conf.ComputeEnv.InstanceRole),
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
	if b.dryRun {
		s, err := json.MarshalIndent(input, "", "  ")
		if err != nil {
			return nil, err
		}
		fmt.Println(string(s))
		return nil, nil
	}
	return batchCli.CreateComputeEnvironment(input)

}

func (b *batchsvc) CreateJobQueue() (*batch.CreateJobQueueOutput, error) {
	batchCli := batch.New(b.sess)

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
		Priority:                aws.Int64(1),
		State:                   aws.String("ENABLED"),
	}

	if b.dryRun {
		s, err := json.MarshalIndent(input, "", "  ")
		if err != nil {
			return nil, err
		}
		fmt.Println(string(s))
		return nil, nil
	}

	return batchCli.CreateJobQueue(input)
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
