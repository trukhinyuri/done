#!/usr/bin/env bash
version=0.1
time=$(date +"%y-%m-%d %T")
echo $version
echo "$time"
env GOOS=darwin go build -o "done" -ldflags="-X 'main.BuildTime=$time' -X 'main.BuildVersion=$version'" .