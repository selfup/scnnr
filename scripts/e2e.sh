#!/usr/bin/env bash

set -eou pipefail

entrypoint=$E2E_DIR

if [[ $E2E_DIR == "" ]]; then
    entrypoint=$HOME
fi

echo "--- FILE SIZE FINDER DRY RUN: BEGIN ---"

go run main.go -m fsf -s 1MB -d $entrypoint

echo "--- FILE SIZE FINDER DRY RUN: DONE ---"

sleep 2

##################################################
##################################################
##################################################

echo "--- FILE NAME FINDER DRY RUN: BEGIN---"

go run main.go -m fnf -f main,DEFCON -p $entrypoint

echo "--- FILE NAME FINDER DRY RUN: DONE ---"

sleep 2

##################################################
##################################################
##################################################

echo "--- SCANNER DRY RUN: BEGIN---"

go run main.go -k main,const,let,var,for -p $entrypoint

echo "--- SCANNER DRY RUN: DONE ---"

sleep 2

echo "--- FILE FINGERPRINT FINDER DRY RUN: BEGIN---"

go run main.go -m fff -k $(go run cmd/checksum/main.go) -d .

echo "--- FILE FINGERPRINT FINDER DRY RUN: DONE---"
