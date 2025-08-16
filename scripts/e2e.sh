#!/usr/bin/env bash

set -eou pipefail

echo "--- FILE SIZE FINDER DRY RUN: BEGIN ---"

go run main.go -m fsf -s 1MB -d $HOME

echo "--- FILE SIZE FINDER DRY RUN: DONE ---"

sleep 2

##################################################
##################################################
##################################################

echo "--- FILE NAME FINDER DRY RUN: BEGIN---"

go run main.go -m fnf -f main,DEFCON -p $HOME

echo "--- FILE NAME FINDER DRY RUN: DONE ---"

sleep 2

##################################################
##################################################
##################################################

echo "--- SCANNER DRY RUN: BEGIN---"

go run main.go -k main,const,let,var,for -p $HOME

echo "--- SCANNER DRY RUN: DONE ---"

sleep 2

echo "--- FILE FINGERPRINT FINDER DRY RUN: BEGIN---"

go run main.go -m fff -k $(go run cmd/checksum/main.go) -d .

echo "--- FILE FINGERPRINT FINDER DRY RUN: DONE---"

sleep 2

echo "--- SCANNER LINE RUN: BEGIN---"

go run main.go -k main,const,let,var,for -p $HOME -l

echo "--- SCANNER LINE RUN: DONE ---"

sleep 2

echo "--- SCANNER LINE AND COL RUN: BEGIN---"

go run main.go -k main,const,let,var,for -p $HOME -c

echo "--- SCANNER LINE AND COL RUN: DONE ---"
