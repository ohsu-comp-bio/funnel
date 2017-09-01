package compute

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// Backend is responsible for submitting a task. For some backends such as HtCondor,
// Slurm, and AWS Batch this amounts to scheduling the task. For others such as
// Openstack this may include provisioning a VM and then running the task.
type Backend interface {
	Submit(*tes.Task) error
}
