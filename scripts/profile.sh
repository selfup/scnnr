#!/usr/bin/env bash

###
# Please get pprof: go get github.com/pkg/profile
# import it
# defer profile.Start().Stop()
# OR
# defer profile.Start(profile.MemProfile).Stop()
# run this script to profile!
# real    0m0.456s
# user    0m1.302s
# sys     0m0.218s
# Max 5.5MB on average for 33k files (node_modules)
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

time go run main.go -dir=artifact -ext=.js -kwd=cache

sleep 1

PROF_PATH="$(ls -d /tmp/profile*)/$PROF_TYPE.pprof"

go tool pprof -http=:8080 $PROF_PATH
