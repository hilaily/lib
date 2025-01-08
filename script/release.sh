#!/bin/bash

dir=$1

cd $dir
ls
go mod tidy
git add .
git diff --quiet HEAD || git commit -m "release $dir" && git push
modtool tag new patch
