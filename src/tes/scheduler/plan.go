package scheduler

type State int
const (
  Accepted State = iota
  Rejected
)

type TaskPlan interface {
  TaskID() string
  WorkerID() string
  State() State
  SetState(State)
  Execute()
  // TODO
  // In the future this could describe to the scheduler
  // the costs/benefits of running this task with this plan,
  // for example, it could convey that there is cached data
  // (data locality is desired).
  // Other ideas:
  // - hard/soft resource requirements. Can this backend offer
  //   higher performance?
  // - monetary cost: maybe this backend knows about the cost
  //   of the cloud instance.
  // - startup time: maybe this backend needs to start a worker
  //   and that could take a few minutes
  // - related task locality: maybe this task is online and related
  //   to others, and a shared, local, fast network would be better.
  // - SLA: maybe this backend can run this, but with the caveat that
  //   the task could be interrupted (AWS spot?). Or maybe this
  //   cluster is prone to having nodes go down.
}

func NewPlan(id string, workerid string) TaskPlan {
  return &basePlan{id, workerid, Proposed}
}

type basePlan struct {
  id string
  workerID string
  state State
}
func (b *basePlan) TaskID() string {
  return b.id
}
func (b *basePlan) WorkerID() string {
  return b.workerID
}
func (b *basePlan) State() State {
  return b.state
}
func (b *basePlan) SetState(s State) {
  b.state = s
}
func (b *basePlan) Execute() {}
