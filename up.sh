#!/bin/bash
# Bash script to keep the studio status program running, even in the event of crashes.
GOPATH=/home/pi/go
cd /home/pi/go/src/github.com/vanderbilt-design-studio/studio_status_go
git pull
go build
while :
do
	(sleep 15; ./studio_status_go) & # Wait 15s, start program and send to background
	sleep 30             # Sleep for 30s
	kill %1              # Kill it
	./studio_status_go 1>>crashes.log 2>>crashes.log # Start it again & log all output
	date >> crashes.log  # Append crash times to log for debugging
done
