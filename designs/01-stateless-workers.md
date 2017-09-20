# Stateless workers

Funnel should not manage persistent workers itself.

A “persistent worker” is a long-lived service/daemon worker process,
which processes multiple tasks tasks over time.
These long-lived processes come with additional complexity:

* Restarting a worker takes care.
* If a worker is restarted, it will lose track of all its running tasks.
* If a worker loses track of a task, it will keep running in Docker.
* Error handling takes care.
* Funnel must watch for unhealthy workers.
* When a worker is dead, it must be deleted from the database.
* Scheduling takes care.
* Funnel must track the available resources of all workers and provide this details to a scheduler process.
* Scheduling code adds complexity and extra code to maintain.

Almost all of Funnel’s backends delegate these responsibilities to external systems;
a worker process is started for each task individually. We call this “worker-per-task”.

* AWS Batch is responsible for worker management.
* The Funnel Cloud proposal is exploring starting a VM for each task in Google Cloud.
* All HPC scheduler backends (HTCondor, Slurm, etc.) are worker-per-task.
  * This makes the most sense in these systems, where a long-lived service process would be frowned
    upon because it would be unfairly holding on to resources.
    In other words, Funnel is taking unfair priority over the HPC scheduler.

The only backend with long-lived workers is the “manual” cluster (which is used in exastack).
This is where we are running into problems, and the impetus for this proposal.

Many advanced systems exist to schedule resources and manage long-lived services:

* Kubernetes
* Mesos
* Docker Swarm
* Nomad
* HPC schedulers (of which there are at least a dozen)

These typically have advanced capability to handle the complexities listed above.
Also, there’s a benefit to having a single system with the responsibility of scheduling and maintaining services,
instead of multiple, different systems with different features and interfaces.

### Considerations

* Reasons to restart a worker
  * New version of Funnel
  * Critical OS security patch
  * Change to VM config, such as updating system service
* If the cluster is busy, restarting a worker is currently not possible.
* It’s not currently possible to mark a worker so that it doesn’t accept any new tasks.
* We want Funnel to be simple to start, and setting up an external cluster scheduler is not simple.
  * On the other hand, most clusters have an existing cluster scheduler.
* Worker working directories need to be cleaned up.
  * In two days we ran out of disk space in the initial deployment in exastack (~250GB).
* Long-lived workers make it impossible for two workers of different versions to share the same VM. This is useful for easy deployment of experimental code.
* Sometimes it’s useful to change worker config for a single task
  * For example, want to configure a task’s runner not to delete its working directory for debugging a failing task, but don’t want to reconfigure/redeploy the whole cluster.
* Fair scheduling of resources is a critical feature and non-trivial to implement
  * Currently, Funnel’s manual cluster is first-come, first-served.
  * For example, if Alex submits 1000 smc-het tasks to the queue which will take a total of 10 days, and then 10 minutes later Michal submits 1000 machine learning tasks which will take in total 1 hour, Michal’s tasks won’t run for 10 days because Alex’s tasks were created first.

### Option 1: delegate
We could decide to remove all support for long-lived workers from Funnel, and delegate that responsibility to other services. If we decide we need something to manage long-lived workers that is not provided by an existing project, that code lives as a separate service.

#### Pros
* Funnel becomes more simple and focused, which ideally leads to better documentation, less bugs, easier deployment, and generally more success.

#### Cons

* More work to deploy a Funnel cluster in an environment with no existing scheduler/service manager.

### Option 2: fix the management issues
In this solution we decide to keep the support for long-lived workers in Funnel and commit to making it a production-worthy cluster manager and scheduler. In order to do that, we need to solve the issues with managing workers.

* Workers need to be restartable.
  * Workers might write their last known state to a local file, and on restart they would resolve that last known state with the current state of the world.
  * Workers need to have a consistent ID across starts.
* Workers need to be capable of being marked so that they don’t receive any new tasks, but finish their existing tasks.
* The server needs to improve it’s health checking of workers.
* The code for maintaining and scheduling to long-lived workers needs to be clearly separated into it’s own component.
* The hard part is that there are likely things I haven’t thought of in this list, which will crop up over time and require more maintenance, work, code, and complexity.

### Option 3: keep the code, but only support development, not production
In this solution we keep the code, but we organize and document it such that it’s clear that it is for testing, development, and experimentation only.

We would probably need to reorganize the code a bit, in order to contain this experimental code into a single package.
Probably move server/scheduler_service.go and most of the scheduler package into something like scheduler/manual

#### Pros

* Allows users to set up a Funnel cluster manually if needed
* Remove responsibility from us to maintain a potentially complex system.
  * Also guides us in making decisions about this part of the code.

#### Cons
* Experimental code in a code base has the potential to leak into the production code, causing extra maintenance, bugs, and even security holes.

#### Related Issues

https://github.com/ohsu-comp-bio/funnel/issues/180  
https://github.com/ohsu-comp-bio/funnel/issues/25  
https://github.com/ohsu-comp-bio/funnel/issues/47  
https://github.com/ohsu-comp-bio/funnel/issues/166  
https://github.com/ohsu-comp-bio/funnel/issues/29  
https://github.com/ohsu-comp-bio/funnel/issues/40  

