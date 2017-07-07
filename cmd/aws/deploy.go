package aws

import (
	"fmt"
)

func deploy(conf Config) error {
	cli := newBatchClient(conf)

	a, aerr := cli.CreateComputeEnvironment()
	fmt.Println(a, aerr)

	b, berr := cli.CreateJobQueue()
	fmt.Println(b, berr)

	c, cerr := cli.CreateJobDef()
	fmt.Println(c, cerr)

	return nil
}
