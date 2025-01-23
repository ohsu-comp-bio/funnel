package elastic

import (
	"encoding/json"
	"fmt"
	"strconv"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/ohsu-comp-bio/funnel/server"
	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/net/context"
)

// Custom unmarshaller where unknown JSON properties do not cause an error.
var customJson = protojson.UnmarshalOptions{
	DiscardUnknown: true,
}

func int64ToStr(i *int64) string {
	return strconv.FormatInt(*i, 10)
}

type TaskOwner struct {
	Owner string `json:"owner"`
}

func (es *Elastic) getTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, string, string, error) {
	g := es.client.Get(es.taskIndex, req.Id)

	switch req.View {
	case tes.View_MINIMAL.String():
		g = g.SourceIncludes_(minimalInclude...)
	case tes.View_BASIC.String():
		g = g.SourceExcludes_(basicExclude...)
	}

	res, err := g.Do(ctx)

	if err != nil {
		return nil, "", "", err
	}

	if !res.Found {
		return nil, "", "", tes.ErrNotFound
	}

	if userInfo := server.GetUser(ctx); !userInfo.CanSeeAllTasks() {
		partial := TaskOwner{}
		_ = json.Unmarshal(res.Source_, &partial)
		if !userInfo.IsAccessible(partial.Owner) {
			return nil, "", "", tes.ErrNotPermitted
		}
	}

	seqNo := int64ToStr(res.SeqNo_)
	primaryTerm := int64ToStr(res.PrimaryTerm_)

	task := tes.Task{}
	err = customJson.Unmarshal(res.Source_, &task)
	return &task, seqNo, primaryTerm, err
}

// GetTask gets a task by ID.
func (es *Elastic) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	task, _, _, err := es.getTask(ctx, req)
	return task, err
}

// ListTasks lists tasks, duh.
func (es *Elastic) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	pageSize := tes.GetPageSize(req.GetPageSize())

	matchFilters := make(map[string]types.MatchQuery)
	termFilters := make(map[string]types.TermQuery)

	if userInfo := server.GetUser(ctx); !userInfo.CanSeeAllTasks() {
		termFilters["owner"] = types.TermQuery{Value: userInfo.Username}
	}

	if req.State != tes.Unknown {
		termFilters["state"] = types.TermQuery{Value: req.State.String()}
	}

	for k, v := range req.GetTags() {
		field := fmt.Sprintf("tags.%s.keyword", k)
		matchFilters[field] = types.MatchQuery{Query: v}
	}

	query := types.Query{
		Bool: &types.BoolQuery{},
	}

	if len(termFilters) > 0 {
		query.Bool.Filter = append(query.Bool.Filter, types.Query{Term: termFilters})
	}
	if len(matchFilters) > 0 {
		query.Bool.Filter = append(query.Bool.Filter, types.Query{Match: matchFilters})
	}

	sort := types.SortOptions{
		SortOptions: map[string]types.FieldSort{
			"id": {Order: &sortorder.Desc},
		},
	}

	search := es.client.Search().
		Index(es.taskIndex).
		Query(&query).
		Size(pageSize).
		Sort(sort).
		ErrorTrace(true)

	if req.PageToken != "" {
		search.SearchAfter(req.PageToken)
	}

	switch req.View {
	case tes.View_BASIC.String():
		search.SourceExcludes_(basicExclude...)
	case tes.View_MINIMAL.String():
		search.SourceIncludes_(minimalInclude...)
	}

	res, err := search.Do(ctx)
	if err != nil {
		return nil, err
	}

	resp := &tes.ListTasksResponse{}
	for i, hit := range res.Hits.Hits {
		task := &tes.Task{}
		err := customJson.Unmarshal(hit.Source_, task)
		if err != nil {
			return nil, err
		}

		if i == pageSize-1 {
			resp.NextPageToken = task.Id
		}

		resp.Tasks = append(resp.Tasks, task)
	}

	return resp, nil
}
