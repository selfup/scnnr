#!/usr/bin/env bash

set -e

echo 'checking remotes for release repo..'

git remote -v | grep release

if [[ $? == '0' ]]
then
  echo 'release set, pushing..'
else
  echo 'release not set, adding ssh release to remotes..'
  git remote add release git@gitlab.com:selfup/scnnr
  git remote -v
  echo 'release set, pushing..'
fi

git push release master
