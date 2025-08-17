#!/usr/bin/env bash

set -eou pipefail

e2e_dir=$HOME

if [[ $ETE_DIR != "" ]]
then
    e2e_dir=$ETE_DIR
fi

##################################################
##################################################
##################################################

echo "--- FILE SIZE FINDER DRY RUN: BEGIN ---"

go run main.go -m fsf -s 1MB -d $e2e_dir

echo "--- FILE SIZE FINDER DRY RUN: DONE ---"

sleep 2

##################################################
##################################################
##################################################

echo "--- FILE NAME FINDER DRY RUN: BEGIN---"

go run main.go -m fnf -f main,DEFCON -p $e2e_dir

echo "--- FILE NAME FINDER DRY RUN: DONE ---"

sleep 2

##################################################
##################################################
##################################################

echo "--- SCANNER DRY RUN: BEGIN---"

go run main.go -k main,const,let,var,for -p $e2e_dir

echo "--- SCANNER DRY RUN: DONE ---"

sleep 2

##################################################
##################################################
##################################################

echo "--- FILE FINGERPRINT FINDER DRY RUN: BEGIN---"

go run main.go -m fff -k $(go run cmd/checksum/main.go) -d .

echo "--- FILE FINGERPRINT FINDER DRY RUN: DONE---"

sleep 2

##################################################
##################################################
##################################################

echo "--- SCANNER LINE RUN: BEGIN---"

go run main.go -k main,const,let,var,for -p $e2e_dir -l

echo "--- SCANNER LINE RUN: DONE ---"

sleep 2

##################################################
##################################################
##################################################

echo "--- SCANNER LINE AND COL RUN: BEGIN---"

go run main.go -k main,const,let,var,for -p $e2e_dir -c

echo "--- SCANNER LINE AND COL RUN: DONE ---"

sleep 2

##################################################
##################################################
##################################################

echo "--- SCANNER DIR EXCLUDE RUN: BEGIN---"

sleep 2

go run main.go -k main,const,let,var,for -p $e2e_dir -c -xd ".git"

echo "--- SCANNER DIR EXCLUDE RUN: DONE ---"

sleep 2

##################################################
##################################################
##################################################

echo "--- SCANNER EXTENSION EXCLUDE RUN: BEGIN---"

go run main.go -k main,const,let,var,for -p $e2e_dir -c -xe ".yml,.sh,.md"

echo "--- SCANNER EXTENSION RUN: DONE ---"

sleep 2

##################################################
##################################################
##################################################

echo "--- SCANNER DIRECTORY AND EXTENSION EXCLUDE RUN: BEGIN---"

go run main.go -k main,const,let,var,for -p $e2e_dir -c -xd ".git" -xe ".yml,.sh,.md"

echo "--- SCANNER DIRECTORY AND EXTENSION RUN: DONE ---"
