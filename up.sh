#!/bin/bash
# Bash script to keep the studio status program running, even in the event of crashes.
cd $HOME/go/src/github.com/vanderbilt-design-studio/studio_status_go
export GOPATH=$HOME/go
git pull
go build
while :
do
	./studio_status_go 2>>crashes.log # Start it & log all errors
	date >> crashes.log  # Append crash times to log for debugging
done
