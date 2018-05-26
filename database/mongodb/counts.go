package mongodb

import (
	"context"

	"github.com/ohsu-comp-bio/funnel/tes"
	"gopkg.in/mgo.v2/bson"
)

type stateCount struct {
	State tes.State `bson:"_id"`
	Count int       `bson:"count"`
}

// TaskStateCounts returns the number of tasks in each state.
func (db *MongoDB) TaskStateCounts(ctx context.Context) (map[string]int32, error) {
	pipe := db.tasks.Pipe([]bson.M{
		{"$group": bson.M{"_id": "$state", "count": bson.M{"$sum": 1}}},
	})

	recs := []stateCount{}
	err := pipe.All(&recs)
	if err != nil {
		return nil, err
	}

	counts := map[string]int32{}
	for _, rec := range recs {
		counts[rec.State.String()] = int32(rec.Count)
	}

	return counts, nil
}
