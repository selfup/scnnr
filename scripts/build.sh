#!/usr/bin/env bash

set -e

mkdir -p bin/mac
mkdir -p bin/linux
mkdir -p bin/windows

if [[ -f main ]]
then
  rm main
fi

if [[ -f main.exe ]]
then
  rm main.exe
fi

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build main.go
mv main bin/mac/scnnr

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
mv main bin/linux/scnnr

CGO_ENABLED=0 GOOS=windows GOARCH=386 go build main.go
mv main.exe bin/windows/scnnr.exe
