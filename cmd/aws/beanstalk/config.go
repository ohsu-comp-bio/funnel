package beanstalk

type beanstalkConfig struct {
	ApplicationName   string
	EnvironmentName   string
	CNAMEPrefix       string
	SolutionStackName string
}

func defaultConfig() beanstalkConfig {
	return beanstalkConfig{
		ApplicationName:   "funnel",
		CNAMEPrefix:       "funnel",
		EnvironmentName:   "funnel",
		SolutionStackName: "64bit Amazon Linux 2017.03 v2.7.4 running Docker 17.03.2-ce",
	}
}
