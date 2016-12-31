package scheduler

func CoordinateScheduler(c Coordinator, s Scheduler) {
  for {
    wl := c.Workload()
    for _, plan := s.Schedule(wl) {
      go func(plan TaskPlan) {
        accepted := c.Submit(plan)
        if accepted {
          plan.Execute()
        }
      }(plan)
    }
  }
}
