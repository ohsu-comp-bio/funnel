package mongodb

import (
	"fmt"

	"github.com/ohsu-comp-bio/funnel/tes"
	"gopkg.in/mgo.v2/bson"
)

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (db *MongoDB) ReadQueue(n int) []*tes.Task {
	var tasks []*tes.Task
	err := db.tasks.Find(bson.M{"state": tes.State_QUEUED}).Sort("creationtime").Select(basicView).Limit(n).All(&tasks)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return tasks
}
