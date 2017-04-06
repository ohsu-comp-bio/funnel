ROOT="$( cd $DIR/../../ && pwd )"

GCE_ERR=0

function gce {
  if [ $GCE_ERR -eq 0 ]; then
    gcloud compute "$@"
    GCE_ERR=$?
  else
    printf "Skipping due to previous error\n"
  fi
}

function gce_always {
  gcloud compute "$@"
}

function log {
  printf "\n======== $1 ========\n\n"
}

function gce_ssh {
  gce ssh $1 --command "$2"
}

function gce_wait_for_ssh {
  START_ERR=0
  for i in {1..5}; do
    gcloud compute ssh $1 --command 'echo' 2> /dev/null
    START_ERR=$?
    if [ $START_ERR -eq 0 ]; then
      break
    else
      sleep 30
    fi
  done
  GCE_ERR=$START_ERR
  if [ $START_ERR -ne 0 ]; then
    log "Couldn't connect to instance"
  fi
}
