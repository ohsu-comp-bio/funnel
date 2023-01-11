package mongodb

import (
	"context"

	"github.com/globalsign/mgo/bson"
	"github.com/ohsu-comp-bio/funnel/tes"
)

type stateCount struct {
	State tes.State `bson:"_id"`
	Count int       `bson:"count"`
}

// TaskStateCounts returns the number of tasks in each state.
func (db *MongoDB) TaskStateCounts(ctx context.Context) (map[string]int32, error) {
	sess := db.sess.Copy()
	defer sess.Close()

	pipe := db.tasks(sess).Pipe([]bson.M{
		{"$sort": bson.M{"state": 1}},
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
