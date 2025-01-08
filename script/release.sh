#!/bin/bash

dir=$1

cd $dir
go mod tidy
go build
git add .
git diff --quiet HEAD || git commit -m "release $dir" && git push
modtool tag new patch
