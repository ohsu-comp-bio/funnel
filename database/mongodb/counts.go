package mongodb

import (
	"context"

	"github.com/ohsu-comp-bio/funnel/tes"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type stateCount struct {
	State tes.State `bson:"_id"`
	Count int       `bson:"count"`
}

// TaskStateCounts returns the number of tasks in each state.
func (db *MongoDB) TaskStateCounts(ctx context.Context) (map[string]int32, error) {
	// sess := db.sess.Copy()
	// defer sess.Close()

	// pipe := db.tasks(sess).Pipe([]bson.M{
	// 	{"$sort": bson.M{"state": 1}},
	// 	{"$group": bson.M{"_id": "$state", "count": bson.M{"$sum": 1}}},
	// })

	// recs := []stateCount{}
	// err := pipe.All(&recs)
	// if err != nil {
	// 	return nil, err
	// }

	// counts := map[string]int32{}
	// for _, rec := range recs {
	// 	counts[rec.State.String()] = int32(rec.Count)
	// }

	// return counts, nil

	sess := db.client
	// defer sess.Disconnect(context.TODO())

	// create group stage
	groupStage := bson.D{
		{"$sort", bson.D{
			{"state", 1},
		}},
		{"$group", bson.D{
			{"_id", "$state"},
			{"count", bson.D{{"$sum", 1}}},
		}}}

	// pass the pipeline to the Aggregate() method
	cursor, err := db.tasks(sess).Aggregate(context.TODO(), mongo.Pipeline{groupStage})
	if err != nil {
		return nil, err
	}

	// display the results
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return nil, nil
}
