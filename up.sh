#!/bin/bash
cd /home/pi/go/src/github.com/vanderbilt-design-studio/studio_status_go
while :
do 
	./studio_status_go &
	sleep 20
	kill %1
	./studio_status_go 1>>crashes.log 2>>crashes.log
	date >> crashes.log
done

