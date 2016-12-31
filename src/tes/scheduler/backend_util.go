package scheduler

type Backend interface {
  Plan(Workload) []TaskPlan
}

func StartBackend(s Scheduler, b Backend) {
  sub := s.Subscribe()

  for wl := range sub.Workload() {
    for _, plan := range b.Plan(wl) {
      go func(plan TaskPlan) {
        resp := s.SubmitPlan(plan)
        if resp == Accepted {
          plan.Execute()
        }
      }(plan)
    }
  }
}
