#!/usr/bin/env bash

export PRELOADED_DB_URL=${PRELOADED_DB_URL:? 'URL to download preloaded db is not set. Exiting.'}
export BHS_DB_PREPARED_DB=false
export BHS_DB_FILE_PATH=${BHS_DB_FILE_PATH:-'./data/blockheaders.db'}
export BHS_DB_PREPARED_DB_FILE_PATH=${BHS_DB_PREPARED_DB_FILE_PATH:-'./data/blockheaders.csv.gz'}
preloaded=false
clean=false

function about() {
    echo "Usage [OPTIONS]

    Starts the block-headers-service application

    Options:
      --preloaded   Load preloaded database if it isn't already loaded
      --clean       Clean database and preloaded db before start of application
      -h  --help    Display this message
    "
}

function clean_db() {
  if [[ -e $BHS_DB_FILE_PATH ]]; then
    echo "Cleaning existing database"
    rm $BHS_DB_FILE_PATH
  fi
  if [[ -e $BHS_DB_PREPARED_DB_FILE_PATH ]]; then
    echo "Cleaning existing preloaded database archive"
    rm $BHS_DB_PREPARED_DB_FILE_PATH
  fi
}

function preload() {
  if [[ -e $BHS_DB_FILE_PATH ]]; then
    echo "There is database file. Skipping preloading database."
    echo "If you want to remove this existing database and use preloaded one, then use the '--clean' argument."
    export BHS_DB_PREPARED_DB=false
  else
    echo "Downloading preloaded database ..."
    wget -nc -O $BHS_DB_PREPARED_DB_FILE_PATH $PRELOADED_DB_URL
    export BHS_DB_PREPARED_DB=true
  fi
}

function start() {
  if $clean ; then
    clean_db
  fi

  if $preloaded ; then
    preload
  fi

  ./"${APP_BINARY}"
}


if test $# -ne 0 ; then
  while test $# -gt 0
  do
      case "$1" in
          --preloaded) preloaded=true
              ;;
          --clean) clean=true
              ;;
          --help) about
              ;;
          -h) about
              ;;
          *) echo "Unknown argument '$1'" && echo "" && about && exit 1
              ;;
      esac
      shift
  done
fi

start
