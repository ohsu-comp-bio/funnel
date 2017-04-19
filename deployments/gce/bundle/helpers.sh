#!/bin/bash

#####################################################
# Helpers

# Check that that a command exists.
check_command() {
  command -v $1 >/dev/null 2>&1
}

log_header() {
  bold=$(tput bold)
  purple=$(tput setaf 171)
  reset=$(tput sgr0)
  printf "\n${bold}${purple}==========  %s  ==========${reset}\n" "$@" 
}

log_error() {
  red=$(tput setaf 1)
  reset=$(tput sgr0)
  printf "${red}%s${reset}\n" "$@"
}

# Skip the command if a previous command failed.
GCE_ERR=0
check_err() {
  if [ $GCE_ERR -eq 0 ]; then
    exec "$@"
    GCE_ERR=$?
  else
    log_error "Skipping due to previous error\n"
  fi
}

# The instance will take some time to boot, so check for a successful
# SSH connection every 30 seconds, for up to 5 minutes.
wait_for_ssh() {
  START_ERR=0
  for i in {1..10}; do
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
    log_error "Error: Couldn't connect to instance\n"
  fi
}
