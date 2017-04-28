---
Demo:
  - Title: Start Funnel
    Cmd: $ funnel server

  - Title: Run a task
    Desc: Returns a task ID.
    Cmd: |
      $ funnel run 'md5sum $src' -c ubuntu --in src=~/src.txt
      b41pkv2rl6qjf441avd0

  - Title: Get the task
    Desc: Returns state, logs, and more.
    Cmd: $ funnel task get b41pkv2rl6qjf441avd0

  - Title: List all the tasks
    Cmd: $ funnel task list

  - Title: View the terminal dashboard
    Cmd: $ funnel dashboard

# - Title: Move to the cloud.
#   Desc: |
#     Google, Amazon, Microsoft, HPC, and more.
#   Cmd: |
#     $ gcloud auth login
#     $ funnel deploy gce
#     $ funnel run 'md5sum'                \
#         --stdin gs://pub/input.txt       \
#         --stdout gs://my-bkt/output.txt

  - Title: Use a remove server
    Cmd: $ funnel run --server http://funnel.example.com ...

  - Title: Example tasks
    Cmd: |
      $ funnel example list
      $ funnel example hello-world

  - Title: Get help
#   Desc: The Funnel CLI is extensive.
    Cmd: $ funnel help

#  - Title: File a bug.
#    Desc: It happens.
#    Cmd: $ funnel bug

  - Title: Get the code
    Cmd: $ go get github.com/ohsu-comp-bio/funnel

#  - Title: Hack together a workflow.
#    Desc: Bash-fu. Hadouken!
#    Cmd: |
#      $ funnel run <<TPL
#      TPL

#  - Title: Use a workflow language.
#    Desc: Level up with CWL and WDL.
    
---

Homepage content is written in layouts/index.html
