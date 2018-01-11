#!/bin/bash
# Bash script to keep the studio status program running, even in the event of crashes.

cd /home/pi/go/src/github.com/vanderbilt-design-studio/studio_status_go
git pull
go build
while :
do 
	./studio_status_go & # Start the program and send it to the background
	sleep 20             # Sleep for 20s
	kill %1              # Kill it
	./studio_status_go 1>>crashes.log 2>>crashes.log # Start it again & log all output
	date >> crashes.log  # Log crash times for debugging
done

