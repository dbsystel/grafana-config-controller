#!/bin/bash

progname=$0
VERBOSITY=0
LIST_ALL=0
EXPORT=0
EXPORT_ALL=0
DATASOURCE=0

: "${GRAFANA_URL:=unknown}"
: "${GRAFANA_DATASOURCTASOURCE=unknown}"
: "${GRAFANA_LOGIN_PASSWORD_FILE:=unknown}"
: "${GRAFANA_LOGIN_USER:=$(whoami)}"
: "${GRAFANA_DATASOURCES_DIRECTORY:=./datasources}"

function usage()
{
   cat << HEREDOC

   Usage: $progname ARGUMENTS

   required arguments:
     -g, --grafana-url URL                      specify the url of the corresponding grafana
                                                (eg. https://my-grafana-url)
     -l, --list                                 specify if you wish to list all datasources in correspondig grafana
     OR                                         OR
     -e, --export                               if you wish to export one datasource with given name (--name)
     -n, --name NAME                            specify datasource name to export
     OR                                         OR
     -a, --all                                  if you wish to export all datasources from corresponding grafana

   optional arguments:
     -h, --help                                 show this help message and exit
     -v, --verbose                              increase the verbosity of the bash script
     -u, --user NAME                            specify grafana login username
                                                (default: whoami)
     -p, --password-file PATH                   specify password file
                                                (default: interactive typein)
     -d, --directory PATH                       specify directory for exporting datasources
                                                (default: ./datasources/ (directory)
   examples:
     ./export-datasources.bash -g https://my-grafana-url -l                       list all datasources
     ./export-datasources.bash -g https://my-grafana-url -e -n prometheus         export one datasource named prometheus
     ./export-datasources.bash -g https://my-grafana-url -e -a                    export all datasources

HEREDOC
}

export_one_datasource() {
    local datasource_saving_name=`echo ${GRAFANA_DATASOURCE_NAME//_/-} | cut -c 1-62 | tr '[A-Z]' '[a-z]'`
    echo "Downloading datasource to: $GRAFANA_DATASOURCES_DIRECTORY/$datasource_saving_name.json"
    datasource_json=$(get_datasource "$GRAFANA_DATASOURCE_NAME")
    num_lines=$(echo "$datasource_json" | wc -l);
    if [ "$num_lines" -le 4 ]; then
      echo "ERROR:
  Couldn't retrieve datasource $GRAFANA_DATASOURCE_NAME! Maybe this datasource does not exist!
      "
      exit 1
    fi
    echo "$datasource_json" >$GRAFANA_DATASOURCES_DIRECTORY/$datasource_saving_name.json
}

export_all_datasources() {
 echo "Starting export of all datasources to: $GRAFANA_DATASOURCES_DIRECTORY"
 local datasources=$(list_datasources)
 local datasource_json
  for datasource in $datasources; do
    local datasource_saving_name=`echo ${datasource//_/-} | cut -c 1-62 | tr '[A-Z]' '[a-z]'`
    echo "Downloading datasource to: $GRAFANA_DATASOURCES_DIRECTORY/$datasource_saving_name.json"
    datasource_json=$(get_datasource "$datasource")
    num_lines=$(echo "$datasource_json" | wc -l);
    if [ "$num_lines" -le 4 ]; then
      echo "ERROR:
  Couldn't retrieve datasource $datasource. Maybe this datasource does not exist!
  Exit
      "
      exit 1
    fi
    echo "$datasource_json" >$GRAFANA_DATASOURCES_DIRECTORY/$datasource_saving_name.json
  done
}

get_datasource() {
  local datasource=$1

  if [[ -z "$datasource" ]]; then
    echo "ERROR:
  A datasource must be specified.
  Exit
  "
    exit 1
  fi
 curl \
    --silent \
    --connect-timeout 10 --max-time 10 \
    --user "$GRAFANA_LOGIN_STRING" \
    $GRAFANA_URL/api/datasources/name/$datasource |
    jq '. | del(.id, .orgId, .version, .readOnly) | .basicAuthPassword=(.basicAuthPassword | @base64) '
}

list_datasources() {
  curl \
    --connect-timeout 10 --max-time 10 \
    --silent \
    --user "$GRAFANA_LOGIN_STRING" \
    $GRAFANA_URL/api/datasources |
    jq -r '.[] | .name' |
    cut -d '/' -f2
  # replace in the future with:
  # jq -r '.[] | select(.type == "dash-db") | .url' |
  # cut -d '/' -f4
}

function prepare() {
echo "Starting..."

if [ "$GRAFANA_LOGIN_PASSWORD_FILE" == "unknown" ]; then
    read -s -p "Please type in password for user $GRAFANA_LOGIN_USER:" GRAFANA_LOGIN_PASSWORD
    echo ""
    : "${GRAFANA_LOGIN_STRING:=$GRAFANA_LOGIN_USER:$GRAFANA_LOGIN_PASSWORD}"
else
    GRAFANA_LOGIN_PASSWORD_FILE_CONTENT=`cat $GRAFANA_LOGIN_PASSWORD_FILE`
    : "${GRAFANA_LOGIN_STRING:=$GRAFANA_LOGIN_USER:$GRAFANA_LOGIN_PASSWORD_FILE_CONTENT}"
fi
[ -d $GRAFANA_DATASOURCES_DIRECTORY ] || mkdir -p $GRAFANA_DATASOURCES_DIRECTORY
}

function test_login() {
 echo "Checking connection and authentication..."
 curl_response=$(curl --connect-timeout 10 --max-time 10 --write-out %{http_code} --silent --user "$GRAFANA_LOGIN_STRING" --output /dev/null $GRAFANA_URL/api/dashboards/home)
 if [ "$curl_response" -eq 200 ] ; then
   echo "Authenticated - OK"
 else
    echo "ERROR:
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
    echo ""
    echo "List of all datasource names of connected grafana:"
    echo ""
        list_datasources
  else
        if [ "$EXPORT_ALL" -gt 0 ]; then
                export_all_datasources
        else
                export_one_datasource
        fi
  fi
}

OPTS=$(getopt -o "g:len:ahvu:p:d:" --long "grafana-url:,list,export,name:,all,help,verbose,user:,password-file:,directory:" -n "$progname" -- "$@")
if [ $? -eq  0 ] ; then
  eval set -- "$OPTS"
  while true; do
    # uncomment the next line to see how shift is working
    # echo "\$1:\"$1\" \$2:\"$2\""
    case "$1" in
      -g | --grafana-url ) GRAFANA_URL=$2; shift 2 ;;
      -l | --list ) LIST_ALL+=1; shift ;;
      -e | --export ) EXPORT+=1; shift ;;
      -n | --name ) GRAFANA_DATASOURCE_NAME=$2; shift 2;;
      -a | --all ) EXPORT_ALL+=1; shift ;;
      -h | --help ) usage; exit 0;;
      -v | --verbose ) VERBOSITY+=1; shift ;;
      -u | --user ) GRAFANA_LOGIN_USER=$2; shift 2 ;;
      -p | --password-file ) GRAFANA_LOGIN_PASSWORD_FILE=$2; shift 2 ;;
      -d | --directory ) GRAFANA_DATASOURCES_DIRECTORY=$2; shift 2 ;;
      -- ) shift; break ;;
      * ) break ;;
    esac
  done

  if [ "$GRAFANA_URL" == "unknown" ] ||
     ( [ $LIST_ALL -eq 0 ] && [ $EXPORT -eq 0 ] ) ||
     ( [ $LIST_ALL -gt 0 ] && [ $EXPORT -gt 0 ] ) ||
     ( [ $LIST_ALL -eq 0 ] && [ $EXPORT_ALL -eq 0 ] && [ "$GRAFANA_DATASOURCE_NAME"  == "unknown" ] ) ;then

     if [ $VERBOSITY -gt 0 ]; then

     cat << DEBUG_OUTPUT

     Debug Output:

     GRAFANA_URL:                         $GRAFANA_URL
     GRAFANA_LOGIN_USER:                  $GRAFANA_LOGIN_USER
     GRAFANA_LOGIN_PASSWORD_FILE:         $GRAFANA_LOGIN_PASSWORD_FILE
     GRAFANA_DATASOURCES_DIRECTORY:       $GRAFANA_DATASOURCES_DIRECTORY
     GRAFANA_DATASOURCE_NAME:             $GRAFANA_DATASOURCE_NAME

DEBUG_OUTPUT
     fi
     usage
     exit 1
  fi
  main
else
  echo "Error in command line arguments." >&2
  usage
fi