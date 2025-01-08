#!/bin/bash

dir=$1

if [ "X$dir" == "X" ]; then
	dir = $(pwd)
fi

cd $dir
go mod tidy
go build
git add .
git diff --quiet HEAD || git commit -m "release $dir" && git push
modtool tag new patch
