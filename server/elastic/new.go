package elastic

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	elastic "gopkg.in/olivere/elastic.v5"
	"time"
)

var minimal = elastic.NewFetchSourceContext(true).Include("id", "state")
var basic = elastic.NewFetchSourceContext(true).
	Exclude("stderr", "stdout", "inputs.content", "system_logs")

// Elastic provides an elasticsearch database server backend.
type Elastic struct {
	client    *elastic.Client
	conf      config.Elastic
	taskIndex string
	nodeIndex string
}

// NewElastic returns a new Elastic instance.
func NewElastic(ctx context.Context, conf config.Elastic) (*Elastic, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(conf.URL),
		elastic.SetSniff(false),
		elastic.SetRetrier(
			elastic.NewBackoffRetrier(
				elastic.NewExponentialBackoff(time.Millisecond*50, time.Minute),
			),
		),
	)
	if err != nil {
		return nil, err
	}
	es := &Elastic{
		client,
		conf,
		conf.IndexPrefix + "-tasks",
		conf.IndexPrefix + "-nodes",
	}
	if err := es.init(ctx); err != nil {
		return nil, err
	}
	return es, nil
}

// Close closes the database client.
func (es *Elastic) Close() error {
	es.client.Stop()
	return nil
}

func (es *Elastic) initIndex(ctx context.Context, name, body string) error {
	exists, err := es.client.
		IndexExists(name).
		Do(ctx)

	if err != nil {
		return err
	} else if !exists {
		if _, err := es.client.CreateIndex(name).Body(body).Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

// init initializing the Elasticsearch indices.
func (es *Elastic) init(ctx context.Context) error {
	taskMappings := `{
    "mappings": {
      "task":{
        "properties":{
          "id": {
            "type": "keyword"
          },
          "state": {
            "type": "keyword"
          },
          "inputs": {
            "type": "nested"
          }
        }
      }
    }
  }`
	if err := es.initIndex(ctx, es.taskIndex, taskMappings); err != nil {
		return err
	}
	if err := es.initIndex(ctx, es.nodeIndex, ""); err != nil {
		return err
	}
	return nil
}
