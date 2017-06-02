package aws

import (
	"fmt"
)

func deploy() error {
	cli := newBatchClient(DefaultConfig())

	a, aerr := cli.CreateComputeEnvironment()
	fmt.Println(a, aerr)

	b, berr := cli.CreateJobQueue()
	fmt.Println(b, berr)

	c, cerr := cli.CreateJobDef()
	fmt.Println(c, cerr)

	return nil
}
