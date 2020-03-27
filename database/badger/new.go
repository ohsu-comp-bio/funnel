package badger

import (
	"fmt"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

// Badger provides a task database based on the Badger embedded database.
type Badger struct {
	db *badger.DB
}

// NewBadger creates a new database instance.
func NewBadger(conf config.Badger) (*Badger, error) {
	err := fsutil.EnsureDir(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("creating database directory: %s", err)
	}
	opts := badger.DefaultOptions(conf.Path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("opening database: %s", err)
	}
	return &Badger{db: db}, nil
}

// Init initializes the database.
func (db *Badger) Init() error {
	return nil
}

var taskKeyPrefix = []byte("tasks")

func taskKey(id string) []byte {
	idb := []byte(id)
	key := make([]byte, 0, len(taskKeyPrefix)+len(idb))
	key = append(key, taskKeyPrefix...)
	key = append(key, idb...)
	return key
}
