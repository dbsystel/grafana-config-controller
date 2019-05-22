#!/bin/bash

progname=$0
VERBOSITY=0
LIST_ALL=0
EXPORT=0
EXPORT_ALL=0

: "${GRAFANA_URL:=unknown}"
: "${GRAFANA_DASHBOARD_NAME:=unknown}"
: "${GRAFANA_LOGIN_PASSWORD_FILE:=unknown}"
: "${GRAFANA_LOGIN_USER:=$(whoami)}"
: "${GRAFANA_DASHBOARDS_DIRECTORY:=./dashboards}"

function usage()
{
   cat >&2 << HEREDOC

   Usage: $progname ARGUMENTS

   required arguments:
     -g, --grafana-url URL                      specify the url of the corresponding grafana 
                                                (eg. https://my-grafana-url)
     -l, --list                                 specify if you wish to list all dashboards in correspondig grafana
     OR                                         OR
     -e, --export                               if you wish to export one dashboard with given name (--name)
     -n, --name NAME                            specify dashboard name to export
     -f, --filename NAME                        specify dashboard name in directory
     -t, --tags                                 if you wish to export all dashboards with given tags ("tag1,tag2,...")
     OR                                         OR
     -a, --all                                  if you wish to export all dashboards from corresponding grafana

   optional arguments:
     -h, --help                                 show this help message and exit
     -v, --verbose                              increase the verbosity of the bash script
     -u, --user NAME                            specify grafana login username
                                                (default: whoami)
     -p, --password-file PATH                   specify password file
                                                (default: interactive typein)
     -d, --directory PATH                       specify directory for exporting dashboards
                                                (default: ./dashboards/ (directory)

   examples:
     * list all dashboards
       ./export-dashboards.bash -g https://my-grafana-url -l 
     * export one dashboard named basic-monitoring 
       ./export-dashboards.bash -g https://my-grafana-url -e -n basic-monitoring
     * export one dashboard named basic-monitoring into file bm.json
       ./export-dashboards.bash -g https://my-grafana-url -e -n basic-monitoring -f bm.json
     * export all dashboards
       ./export-dashboards.bash -g https://my-grafana-url -e -a

HEREDOC
}

log() {
	echo >&2 $*
}

export_one_dashboard() {
    local dashboard_saving_name=${GRAFANA_DASHBOARD_FILENAME:=$(echo ${GRAFANA_DASHBOARD_NAME//_/-} | cut -c 1-62)}
    log "Downloading dashboard to: $GRAFANA_DASHBOARDS_DIRECTORY/$dashboard_saving_name.json"
    echo "$GRAFANA_DASHBOARDS_DIRECTORY/$dashboard_saving_name.json"
    dashboard_json=$(get_dashboard "$GRAFANA_DASHBOARD_NAME")
    num_lines=$(echo "$dashboard_json" | wc -l);
    if [ "$num_lines" -le 4 ]; then
      log "ERROR:
  Couldn't retrieve dashboard $GRAFANA_DASHBOARD_NAME! Maybe this dashboard does not exist!
      "
      exit 1
    fi
    echo "$dashboard_json" >$GRAFANA_DASHBOARDS_DIRECTORY/$dashboard_saving_name.json
}

export_all_dashboards() {
 log "Starting export of all dashboards to: $GRAFANA_DASHBOARDS_DIRECTORY"
 local dashboards=$(list_dashboards)
 local dashboard_json
  for dashboard in $dashboards; do
    local dashboard_saving_name=`echo ${dashboard//_/-} | cut -c 1-62`
    log "Downloading dashboard to: $GRAFANA_DASHBOARDS_DIRECTORY/$dashboard_saving_name.json"
    dashboard_json=$(get_dashboard "$dashboard")
    num_lines=$(echo "$dashboard_json" | wc -l);
    if [ "$num_lines" -le 4 ]; then
      log "ERROR:
  Couldn't retrieve dashboard $dashboard. Maybe this dashboard does not exist!
  Exit
      "
      exit 1
    fi
    echo "$dashboard_json" >$GRAFANA_DASHBOARDS_DIRECTORY/$dashboard_saving_name.json
  done
}

get_dashboard() {
  local dashboard=$1

  if [[ -z "$dashboard" ]]; then
    log "ERROR:
  A dashboard must be specified.
  Exit
  "
    exit 1
  fi
 curl \
    --silent \
    --connect-timeout 10 --max-time 10 \
    --user "$GRAFANA_LOGIN_STRING" \
    $GRAFANA_URL/api/dashboards/db/$dashboard |
    jq '.dashboard | .id = null | .version = null' 
}

list_dashboards() {
  local tag_string="?"
  for tag in ${GRAFANA_DASHBOARD_TAGS//,/ } ; do
    if [[ "$tag_string" == "?" ]] ; then
      tag_string+="tag=$tag"
    else
      tag_string+="&tag=$tag"
    fi
  done
  if [[ "$tag_string" == "?" ]] ; then tag_string="" ; fi

  curl \
    --connect-timeout 10 --max-time 10 \
    --silent \
    --user "$GRAFANA_LOGIN_STRING" \
    $GRAFANA_URL/api/search$tag_string |
    jq -r '.[] | select(.type == "dash-db") | .uri' |
    cut -d '/' -f2
  # replace in the future with:
  # jq -r '.[] | select(.type == "dash-db") | .url' |
  # cut -d '/' -f4
}

function prepare() {
log "Starting..."

if [ "$GRAFANA_LOGIN_PASSWORD_FILE" == "unknown" ]; then
    read -s -p "Please type in password for user $GRAFANA_LOGIN_USER:" GRAFANA_LOGIN_PASSWORD
    echo ""
    : "${GRAFANA_LOGIN_STRING:=$GRAFANA_LOGIN_USER:$GRAFANA_LOGIN_PASSWORD}"
else
    GRAFANA_LOGIN_PASSWORD_FILE_CONTENT=`cat $GRAFANA_LOGIN_PASSWORD_FILE`
    : "${GRAFANA_LOGIN_STRING:=$GRAFANA_LOGIN_USER:$GRAFANA_LOGIN_PASSWORD_FILE_CONTENT}"
fi
[ -d $GRAFANA_DASHBOARDS_DIRECTORY ] || mkdir -p $GRAFANA_DASHBOARDS_DIRECTORY
}

function test_login() {
 log "Checking connection and authentication..."
 curl_response=$(curl --connect-timeout 10 --max-time 10 --write-out %{http_code} --silent --user "$GRAFANA_LOGIN_STRING" --output /dev/null $GRAFANA_URL/api/dashboards/home)
 if [ "$curl_response" -eq 200 ] ; then
   log "Authenticated - OK"
 else
    log "ERROR:
   Received http_code: $curl_response
   Exit
   "
   exit 1
 fi
}

function main() {
  prepare
  test_login
  if [ "$LIST_ALL" -gt 0 ]; then
    log ""
    log "List of all dashboard names of connected grafana:"
    log ""
  	list_dashboards
  else
  	if [ "$EXPORT_ALL" -gt 0 ]; then
  		export_all_dashboards
  	else
  		export_one_dashboard
  	fi
  fi
}

OPTS=$(getopt -o "g:len:f:ahvu:p:d:t:" --long "grafana-url:,list,export,name:,filename:,all,help,verbose,user:,password-file:,directory:,tags:" -n "$progname" -- "$@")
if [ $? -eq  0 ] ; then
  eval set -- "$OPTS"
  while true; do
    # uncomment the next line to see how shift is working
    # echo "\$1:\"$1\" \$2:\"$2\""
    case "$1" in
      -g | --grafana-url ) GRAFANA_URL=${2%/}; shift 2 ;;
      -l | --list ) LIST_ALL+=1; shift ;;
      -e | --export ) EXPORT+=1; shift ;;
      -n | --name ) GRAFANA_DASHBOARD_NAME=$2; shift 2;;
      -t | --tags ) GRAFANA_DASHBOARD_TAGS=$2; shift 2;;
      -f | --filename ) GRAFANA_DASHBOARD_FILENAME=$2; shift 2;;
      -a | --all ) EXPORT_ALL+=1; shift ;;
      -h | --help ) usage; exit 0;;
      -v | --verbose ) VERBOSITY+=1; shift ;;
      -u | --user ) GRAFANA_LOGIN_USER=$2; shift 2 ;;
      -p | --password-file ) GRAFANA_LOGIN_PASSWORD_FILE=$2; shift 2 ;;
      -d | --directory ) GRAFANA_DASHBOARDS_DIRECTORY=$2; shift 2 ;;
      -- ) shift; break ;;
      * ) break ;;
    esac
  done
  
  if [ "$GRAFANA_URL" == "unknown" ] ||
     ( [ $LIST_ALL -eq 0 ] && [ $EXPORT -eq 0 ] ) ||
     ( [ $LIST_ALL -gt 0 ] && [ $EXPORT -gt 0 ] ) ||
     ( [ $LIST_ALL -eq 0 ] && [ $EXPORT_ALL -eq 0 ] && [ "$GRAFANA_DASHBOARD_NAME"  == "unknown" ] ) ;then

     if [ $VERBOSITY -gt 0 ]; then

     cat << DEBUG_OUTPUT

     Debug Output:

     GRAFANA_URL:                         $GRAFANA_URL
     GRAFANA_LOGIN_USER:                  $GRAFANA_LOGIN_USER
     GRAFANA_LOGIN_PASSWORD_FILE:         $GRAFANA_LOGIN_PASSWORD_FILE
     GRAFANA_DASHBOARDS_DIRECTORY:        $GRAFANA_DASHBOARDS_DIRECTORY
     GRAFANA_DASHBOARD_NAME:              $GRAFANA_DASHBOARD_NAME
     GRAFANA_DASHBOARD_TAGS:              $GRAFANA_DASHBOARD_TAGS

DEBUG_OUTPUT
     fi
     usage
     exit 1
  fi 
  main
else
  log "Error in command line arguments." >&2
  usage
fi