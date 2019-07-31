#!/usr/bin/env bash

if [[ -d scnnr_bins ]]
then
  rm -rf scnnr_bins
fi

if [[ -f scnnr_bins.zip ]]
then
  rm scnnr_bins.zip
fi

./scripts/build.sh

zip -r scnnr_bins.zip scnnr_bins README.md LICENSE
