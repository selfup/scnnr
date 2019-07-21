#!/usr/bin/env bash

###
# go get github.com/pkg/profile
# import "github.com/pkg/profile"
# THEN in main()
# defer profile.Start().Stop()
# OR
# defer profile.Start(profile.MemProfile).Stop()
###

set -e

rm -rf /tmp/profile*

if [[ $1 == "" ]]
then
  echo 'please pass -m for mem.pprof or -c for cpu.pprof'
  exit 1
fi

if [[ $1 == "-m" ]]
then
  PROF_TYPE="mem"
else
  PROF_TYPE="cpu"
fi

time go run main.go -d=. -e=.md -k=cache > .results

sleep 1

PROF_PATH="$(ls -d /tmp/profile*)/$PROF_TYPE.pprof"

go tool pprof -http=:8080 $PROF_PATH
