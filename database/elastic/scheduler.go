package elastic

import (
	"bytes"
	"context"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/tes"
	elastic "gopkg.in/olivere/elastic.v5"
)

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (es *Elastic) ReadQueue(n int) []*tes.Task {
	ctx := context.Background()

	q := elastic.NewTermQuery("state", tes.State_QUEUED.String())
	res, err := es.client.Search().
		Index(es.taskIndex).
		Type("task").
		Size(n).
		Sort("id", true).
		Query(q).
		Do(ctx)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var tasks []*tes.Task
	for _, hit := range res.Hits.Hits {
		t := &tes.Task{}
		err := jsonpb.Unmarshal(bytes.NewReader(*hit.Source), t)
		if err != nil {
			continue
		}

		t = t.GetBasicView()
		tasks = append(tasks, t)
	}

	return tasks
}
