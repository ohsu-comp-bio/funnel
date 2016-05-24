
package ga4gh_task

import (
	"golang.org/x/net/context"
	"ga4gh-tasks"
	"github.com/boltdb/bolt"
	uuid "github.com/nu7hatch/gouuid"
	proto "github.com/golang/protobuf/proto"
	"fmt"
	"log"
	"ga4gh-server/proto"
)


var TASK_BUCKET = []byte("tasks")
var TASKOP_BUCKET = []byte("taskops")
var TASKOP_MAP_BUCKET = []byte("taskop-map")

var TASKOP_ACTIVE = []byte("taskop-active")
var TASKOP_COMPLETE = []byte("taskop-complete")

var TASKOP_LOG = []byte("taskop-log")

type TaskBolt struct {
	db *bolt.DB
}


func NewTaskBolt(path string) *TaskBolt {
	db, _ := bolt.Open(path, 0600, nil)
	//Check to make sure all the required buckets have been created
	db.Update(func(tx *bolt.Tx) error {
		if (tx.Bucket(TASK_BUCKET) == nil) {
			tx.CreateBucket(TASK_BUCKET)
		}
		if (tx.Bucket(TASKOP_BUCKET) == nil) {
			tx.CreateBucket(TASKOP_BUCKET)
		}
		if (tx.Bucket(TASKOP_MAP_BUCKET) == nil) {
			tx.CreateBucket(TASKOP_MAP_BUCKET)
		}
		if (tx.Bucket(TASKOP_ACTIVE) == nil) {
			tx.CreateBucket(TASKOP_ACTIVE)
		}
		if (tx.Bucket(TASKOP_COMPLETE) == nil) {
			tx.CreateBucket(TASKOP_COMPLETE)
		}
		if (tx.Bucket(TASKOP_LOG) == nil) {
			tx.CreateBucket(TASKOP_LOG)
		}
		return nil
	})
	return &TaskBolt{db:db}
}


func (self *TaskBolt) CreateTask(ctx context.Context, task *ga4gh_task_exec.Task) (*ga4gh_task_exec.Task, error) {
	u, _ := uuid.NewV4()
	task.TaskId = u.String()
	log.Println("Adding task", task)
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(TASK_BUCKET)
		v, _ := proto.Marshal(task)
		err := b.Put([]byte(task.TaskId), v )
		return err
	})
	return task, err
}

// / Delete a task
func (self *TaskBolt) DeleteTask(ctx context.Context, task *ga4gh_task_exec.TaskId) (*ga4gh_task_exec.TaskId, error) {
	err := self.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(TASK_BUCKET)
		err := b.Delete([]byte(task.Value))
		return err
	})
	return task, err
}

// / Get a task by its ID
func (self *TaskBolt) GetTask(ctx context.Context, task *ga4gh_task_exec.TaskId) (*ga4gh_task_exec.Task, error) {
	log.Println("Getting Task", task.Value)
	ch := make(chan *ga4gh_task_exec.Task, 1)
	err := self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(TASK_BUCKET)
		v := b.Get([]byte(task.Value))
		if (v == nil) {
			ch <- nil
			return fmt.Errorf("Missing Key %s", task.Value)
		}
		out := ga4gh_task_exec.Task{}
		proto.Unmarshal(v, &out)
		ch <- &out
		return nil
	})
	a := <- ch
	return a, err
}

// / Run a task
func (self *TaskBolt) RunTask(ctx context.Context, request *ga4gh_task_exec.TaskRunRequest) (*ga4gh_task_exec.TaskOpId, error) {
	log.Println("Recieving Task for Queue", request)
	ch := make(chan *ga4gh_task_exec.TaskOpId, 1)
	err := self.db.Update(func(tx *bolt.Tx) error {
		taskopId, _ := uuid.NewV4()

		//find the task info
		var task *ga4gh_task_exec.Task = nil
		if request.EphemeralTask != nil {
			task = request.EphemeralTask
		}
		if len(request.TaskId) > 0 {
			task_b := tx.Bucket(TASK_BUCKET)
			v := task_b.Get([]byte(request.TaskId))
			if v != nil {
				task := &ga4gh_task_exec.Task{}
				proto.Unmarshal(v, task)
			}
		}

		//make sure task op record is completely filled out
		if (task == nil) {
			ch <- nil
			return fmt.Errorf("Task Description not found")
		}
		if (len(task.Docker) == 0) {
			ch <- nil
			return fmt.Errorf("No docker commands found")
		}

		for _, input := range task.GetInputParameters() {
			_, ok := request.TaskArgs.Inputs[input.Name]; if !ok {
				ch <- nil
				return fmt.Errorf("Missing Input:%s", input.Name)
			}
			disk_found := false
			for _, res := range request.EphemeralTask.Resources.Disks {
				if res.Name == input.LocalCopy.Disk {
					disk_found = true
				}
			}
			if !disk_found {
				return fmt.Errorf("Required disk %s not found in resources", input.LocalCopy.Disk)
			}
		}


		taskop := ga4gh_task_exec.TaskOp{
			TaskArgs: request.TaskArgs,
			TaskOpId: taskopId.String(),
			State:ga4gh_task_exec.State_Queued,
			Task: task,
		}
		taskop_b := tx.Bucket(TASKOP_BUCKET)
		v, _ := proto.Marshal(&taskop)
		taskop_b.Put( []byte(taskopId.String()), v )

		active_b := tx.Bucket(TASKOP_ACTIVE)
		active_b.Put([]byte(taskopId.String()), []byte(ga4gh_task_exec.State_Queued.String()))

		ch <- &ga4gh_task_exec.TaskOpId{Value:taskopId.String()}
		return nil
	})
	if err != nil {
		return nil, err
	}
	a := <- ch
	return a, err
}


func (self *TaskBolt) ListTask(ctx context.Context, in *ga4gh_task_exec.TaskListRequest) (*ga4gh_task_exec.TaskListResponse, error) {
	log.Println("Listing Task")

	ch := make(chan *ga4gh_task_exec.Task, 1)

	go self.db.View(func(tx *bolt.Tx) error {
		task_b := tx.Bucket(TASK_BUCKET)
		c := task_b.Cursor()
		log.Println("Scanning")
		for k, v := c.First(); k != nil; k, v = c.Next() {
			out := ga4gh_task_exec.Task{}
			proto.Unmarshal(v, &out)
			ch <- &out
		}
		close(ch)
		return nil
	})

	task_array := make([]*ga4gh_task_exec.Task, 0, 10)
	for t := range ch {
		task_array = append(task_array, t)
	}

	out := ga4gh_task_exec.TaskListResponse{
		Tasks: task_array,
	}

	return &out, nil
}


func (self *TaskBolt) updateTaskOp(taskop *ga4gh_task_exec.TaskOp) (error) {
	self.db.View(func(tx *bolt.Tx) error {
		b_a := tx.Bucket(TASKOP_ACTIVE)
		b_c := tx.Bucket(TASKOP_COMPLETE)
		b_l := tx.Bucket(TASKOP_LOG)
		//if its active
		if v := b_a.Get([]byte(taskop.TaskOpId)); v != nil {
			taskop.State = ga4gh_task_exec.State(ga4gh_task_exec.State_value[string(v)])
		}
		//if its complete
		if v := b_c.Get([]byte(taskop.TaskOpId)); v != nil {
			taskop.State = ga4gh_task_exec.State(ga4gh_task_exec.State_value[string(v)])
		}
		//if there is logging info
		out := make([]*ga4gh_task_exec.TaskOpLog, len(taskop.Task.Docker), len(taskop.Task.Docker))
		for i := range taskop.Task.Docker {
			o := b_l.Get([]byte(fmt.Sprint("%s-%d", taskop.TaskOpId, i)))
			if o != nil {
				var log ga4gh_task_exec.TaskOpLog
				proto.Unmarshal(o, &log)
				out[i] = &log
			} else {
				out[i] = &ga4gh_task_exec.TaskOpLog{}
			}
		}
		taskop.Logs = out
		return nil
	})
	return nil
}

// / Get info about a running task
func (self *TaskBolt) GetTaskOp(ctx context.Context, taskop *ga4gh_task_exec.TaskOpId) (*ga4gh_task_exec.TaskOp, error) {
	log.Printf("Getting Task Info")
	ch := make(chan *ga4gh_task_exec.TaskOp, 1)
	self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(TASKOP_BUCKET)
		v := b.Get([]byte(taskop.Value))
		out := ga4gh_task_exec.TaskOp{}
		proto.Unmarshal(v, &out)
		ch <- &out
		return nil
	})
	a := <- ch
	self.updateTaskOp(a)
	return a, nil
}

func (self *TaskBolt) ListTaskOp(ctx context.Context, in *ga4gh_task_exec.TaskOpListRequest) (*ga4gh_task_exec.TaskOpListResponse, error) {
	log.Printf("Getting Task List")
	ch := make(chan *ga4gh_task_exec.TaskOp, 1)
	go self.db.View(func(tx *bolt.Tx) error {
		taskop_b := tx.Bucket(TASKOP_BUCKET)
		c := taskop_b.Cursor()
		log.Println("Scanning")
		for k, v := c.First(); k != nil; k, v = c.Next() {
			out := ga4gh_task_exec.TaskOp{}
			proto.Unmarshal(v, &out)
			ch <- &out
		}
		close(ch)
		return nil
	})

	task_array := make([]*ga4gh_task_exec.TaskOp, 0, 10)
	for t := range ch {
		self.updateTaskOp(t)
		task_array = append(task_array, t)
	}

	out := ga4gh_task_exec.TaskOpListResponse{
		TasksOps: task_array,
	}
	fmt.Println("Returning", out)
	return &out, nil

}

// / Cancel a running task
func (self *TaskBolt) CancelTaskOp(ctx context.Context, taskop *ga4gh_task_exec.TaskOpId) (*ga4gh_task_exec.TaskOpId, error) {
	self.db.Update(func(tx *bolt.Tx) error {
		b_a := tx.Bucket(TASKOP_ACTIVE)
		b_a.Delete([]byte(taskop.Value))

		b_c := tx.Bucket(TASKOP_COMPLETE)
		b_c.Put([]byte(taskop.Value), []byte(ga4gh_task_exec.State_Canceled.String()))
		return nil
	})
	return taskop, nil
}


func (self *TaskBolt) GetJobToRun(ctx context.Context, request *ga4gh_task_ref.JobRequest) (*ga4gh_task_ref.JobResult, error) {
	log.Printf("Job Request")
	ch := make(chan *ga4gh_task_exec.TaskOp, 1)

	self.db.Update(func (tx *bolt.Tx) error {
		b := tx.Bucket(TASKOP_ACTIVE)
		b_op := tx.Bucket(TASKOP_BUCKET)

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			if (string(v) == ga4gh_task_exec.State_Queued.String()) {
				log.Printf("Found queued job")
				v := b_op.Get(k)
				out := ga4gh_task_exec.TaskOp{}
				proto.Unmarshal(v, &out)
				ch <- &out
				b.Put(k, []byte(ga4gh_task_exec.State_Running.String()))
				return nil
			}
		}
		ch <- nil
		return nil
	})
	a := <- ch
	return &ga4gh_task_ref.JobResult{Task:a}, nil
}

func (self *TaskBolt) UpdateTaskOpStatus(ctx context.Context, stat *ga4gh_task_ref.UpdateStatusRequest) (*ga4gh_task_exec.TaskOpId, error) {
	log.Printf("Set op status")
	self.db.Update(func(tx *bolt.Tx) error {
		ba := tx.Bucket(TASKOP_ACTIVE)
		bc := tx.Bucket(TASKOP_COMPLETE)
		b_l := tx.Bucket(TASKOP_LOG)

		if stat.Log != nil {
			log.Printf("Logging stdout:%s", stat.Log.Stdout )
			d, _ := proto.Marshal(stat.Log)
			b_l.Put([]byte(fmt.Sprint("%s-%d", stat.Id, stat.Step)), d )
		}

		switch stat.State {
		case ga4gh_task_exec.State_Complete, ga4gh_task_exec.State_Error:
			ba.Delete([]byte(stat.Id))
			bc.Put([]byte(stat.Id), []byte(stat.State.String()))
		}
		return nil
	})
	return &ga4gh_task_exec.TaskOpId{Value:stat.Id}, nil
}

func (self *TaskBolt) WorkerPing(ctx context.Context, info *ga4gh_task_ref.WorkerInfo) (*ga4gh_task_ref.WorkerInfo, error) {
	log.Printf("Worker Ping")
	return info, nil
}