#!/bin/bash

set +eux

# shellcheck disable=SC1073
echo "Running for $1 times, sleeping for $2"
for i in $(seq 1 $1); do
  echo "#$i"; sleep $2;
done
