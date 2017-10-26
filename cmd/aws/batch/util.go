package batch

import "github.com/aws/aws-sdk-go/service/batch"

type byRevision []*batch.JobDefinition

func (a byRevision) Len() int {
	return len(a)
}

func (a byRevision) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a byRevision) Less(i, j int) bool {
	return *a[i].Revision > *a[j].Revision
}
