---
Demo:
  - Title: Start Funnel
    Desc: Start simple.
    Cmd: $ funnel server

  - Title: Run a task
    Desc: Returns a task ID.
    Cmd: $ funnel run 'md5sum' --stdin ~/input.txt --stdout ~/input-md5.txt

  - Title: Get the task
    Desc: Returns state, logs, and more.
    Cmd: $ funnel tasks get [task-id]

  - Title: Run lots of tasks
    Desc: That's what clusters are for.
    Cmd: $ funnel run 'md5sum' 

  - Title: List all the tasks
    Desc: TODO
    Cmd: $ funnel tasks list

  - Title: Move to the cloud
    Desc: |
      Google, Amazon, Microsoft, HPC, and more.
    Cmd: |
      $ gcloud auth login
      $ funnel deploy gce
      $ funnel run 'md5sum'                \
          --stdin gs://pub/input.txt       \
          --stdout gs://my-bkt/output.txt

  - Title: Hand-craft a task.
    Desc: Get into the details with JSON or YAML.
    Cmd: |
      $ funnel task create --yaml <<TASK
        name: My Task
        description: Taking Funnel for a drive.
        inputs:
      TASK

  - Title: Get help
    Desc: The Funnel CLI is extensive.
    Cmd: $ funnel help

  - Title: File a bug
    Desc: It happens.
    Cmd: $ funnel bug

  - Title: Get the code
    Desc: Go get it.
    Cmd: $ go get github.com/ohsu-comp-bio/funnel

#  - Title: Hack together a workflow
#    Desc: Bash-fu. Hadouken!
#    Cmd: |
#      $ funnel run <<TPL
#      TPL

#  - Title: Use a workflow language
#    Desc: Level up with CWL and WDL.
    
---

Homepage content is written in layouts/index.html
