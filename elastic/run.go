package main

import (
  "context"
  "fmt"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

func main() {
  ctx := context.Background()
  c := DefaultConfig()
  es, err := NewElastic(c)
  if err != nil {
    panic(err)
  }

  err = es.Init(ctx)

  if err != nil {
    panic(err)
  }

  id := "b76l27irl6qmadfm7dm0"
  task, err := es.GetTask(ctx, id)
  if err != nil {
    panic(err)
  }
  fmt.Println("task", task)

  /*
  id := util.GenTaskID()
  err = es.CreateTask(ctx, &tes.Task{
    Id: id,
    Executors: []*tes.Executor{
      {
        Cmd: []string{"echo", "hello world"},
      },
    },
  })
  if err != nil {
    panic(err)
  }
  */

  ev := events.NewState(id, 0, tes.State_QUEUED)
  err = es.Write(ev)
  if err != nil {
    panic(err)
  }

  tasks, err := es.ListTasks(ctx)
  if err != nil {
    panic(err)
  }

  for _, task := range tasks {
    fmt.Println("tasks", task)
  }

}
