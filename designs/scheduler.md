# Scheduler design

### Considerations

- Multiple cluster scheduling via a single Funnel server
- Each of multiple clusters has a different scheduling policy

### Questions

- What is responsible for detecting dead workers? Does this implementation change per scheduler type?
- Do all scheduler backends share a single worker database?

### Pain points

- possible the implementation is currently buggy, in that multiple calls to schedule happen before a call to scale, so the scheduler backend might over-scale the workers.
