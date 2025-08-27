#!/bin/bash

dirs=(api docker general services)

today=$(date +%F)  # %F gives YYYY-MM-DD
commit_prefix="$today Mirror Update for"

for dir in ${dirs[@]}; do
  for subdir in $dir/*/; do
    msg="$commit_prefix $subdir"
    git add "$subdir"
    git commit -m "$msg"
  done
done

