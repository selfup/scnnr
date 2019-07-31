#!/usr/bin/env bash

set -e

mkdir -p scnnr_bins/mac
mkdir -p scnnr_bins/linux
mkdir -p scnnr_bins/windows

if [[ -f main ]]
then
  rm main
fi

if [[ -f main.exe ]]
then
  rm main.exe
fi

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build main.go
mv main scnnr_bins/mac/scnnr

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
mv main scnnr_bins/linux/scnnr

CGO_ENABLED=0 GOOS=windows GOARCH=386 go build main.go
mv main.exe scnnr_bins/windows/scnnr.exe

chmod +x scnnr_bins/linux/scnnr scnnr_bins/mac/scnnr

if [[ $CI == 'true' ]]
then
  if [[ $VERSION != "" ]]
  then
    echo $VERSION > scnnr_bins/version
  fi
fi
