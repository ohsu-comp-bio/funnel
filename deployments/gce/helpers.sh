DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT="$( cd $DIR/../../ && pwd )"

#Set Colors
#

bold=$(tput bold)
reset=$(tput sgr0)
purple=$(tput setaf 171)
red=$(tput setaf 1)
green=$(tput setaf 76)
blue=$(tput setaf 38)

#
# Headers and  Logging
#

log_header() {
  printf "\n${bold}${purple}==========  %s  ==========${reset}\n" "$@" 
}
log_success() {
  printf "${green}✔ %s${reset}\n" "$@"
}
log_error() {
  printf "${red}✖ %s${reset}\n" "$@"
}
log() {
  printf "${blue}%s${reset}\n" "$@"
}

GCE_ERR=0

gce() {
  if [ $GCE_ERR -eq 0 ]; then
    gcloud compute "$@"
    GCE_ERR=$?
  else
    printf "Skipping due to previous error\n"
  fi
}

gce_always() {
  gcloud compute "$@"
}

check_command() {
  command -v $1 >/dev/null 2>&1
}

gce_ssh() {
  gce ssh $1 --command "$2"
}

gce_wait_for_ssh() {
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
    log_error "Couldn't connect to instance"
  fi
}
