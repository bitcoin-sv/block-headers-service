#!/usr/bin/env bash

export PRELOADED_DB_URL=${PRELOADED_DB_URL:? 'URL to download preloaded db is not set. Exiting.'}
export DB_PREPAREDDB=false
export DB_DBFILEPATH=${DB_DBFILEPATH:-'./data/blockheaders.db'}
export DB_PREPAREDDBFILEPATH=${DB_PREPAREDDBFILEPATH:-'./data/blockheaders.xz'}
preloaded=false
clean=false

function about() {
    echo "Usage [OPTIONS]

    Starts the pulse application

    Options:
      --preloaded   Load preloaded database if it isn't already loaded
      --clean       Clean database and preloaded db before start of application
      -h  --help    Display this message
    "
}

function clean_db() {
  if [[ -e $DB_DBFILEPATH ]]; then
    echo "Cleaning existing database"
    rm $DB_DBFILEPATH
  fi
  if [[ -e $DB_PREPAREDDBFILEPATH ]]; then
    echo "Cleaning existing preloaded database archive"
    rm $DB_PREPAREDDBFILEPATH
  fi
}

function preload() {
  if [[ -e $DB_DBFILEPATH ]]; then
    echo "There is database file. Skipping preloading database."
    echo "If you want to remove this existing database and use preloaded one, then use the '--clean' argument."
    export DB_PREPAREDDB=false
  else
    echo "Downloading preloaded database ..."
    wget -nc -O $DB_PREPAREDDBFILEPATH $PRELOADED_DB_URL
    export DB_PREPAREDDB=true
  fi
}

function start() {
  if $clean ; then
    clean_db
  fi

  if $preloaded ; then
    preload
  fi

  ./pulse
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
