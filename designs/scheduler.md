# Scheduler design

### Considerations

- Multiple cluster scheduling via a single Funnel server
- Each of multiple clusters has a different scheduling policy

### Questions

- What is responsible for detecting dead workers? Does this implementation change per scheduler type?
- Do all scheduler backends share a single worker database?

### Pain points

- possible the implementation is currently buggy, in that multiple calls to schedule happen before a call to scale, so the scheduler backend might over-scale the workers.
  - actually, that's wrong, but the reason is subtle. this bug would only exist if Schedule() was called concurrently.
  - workers are added to the database by AssignTask, and each call to Schedule typically gets a fresh list of workers.
  - this is why Schedule() and Scale() are separate.
