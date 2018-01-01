# Studio Status
### *Golang program for the Studio Status screen in the Design Studio window*

## Setup
1. Follow the instructions [here](https://www.raspberrypi.org/documentation/installation/installing-images/) to install the Raspberry PI OS on a micro SD card. **Install the headless version**
2. Once installed, put that micro sd into the PI. There's a slot, make sure you put it in the right way and don't just jam it in.
3. Plug the Raspberry Pi into the monitor + ethernet + a keyboard.
4. It should do some fancy boot stuff. A login prompt will appear. Log in with username: pi, and password: raspberry.
5. Once logged in, do the following commands:
   
   `sudo apt-get update`
   
   `sudo rpi-update`
   
   `reboot` _Wait until it finishes rebooting, log in again, then continue._
   
   `sudo apt-get install git golang` type y if it asks if you're sure you want to install
   
   `go get github.com/vanderbilt-design-studio/studio_status_go` _If there are any errors, email me. My email is on my Github profile_
   
   `cd ~/go/src/github.com/vanderbilt-design-studio/studio_status_go`
   
   `go build`
   
   `crontab -e` (an editor will pop up, don't worry, I'll guide you through that)
   
   make sure you are on a fresh line, then type the following: `@reboot /home/pi/go/src/github.com/vanderbilt-design-studio/studio_status_go/up.sh`
   
   do Ctrl+ O (not 0) then hit enter
   
   do Ctrl+ X
   
   done! Just restart the Pi now by doing `reboot`
   
## Troubleshooting

### Switch isn't working
   Make sure you have BCM GPIO 17 and 27 (http://pinout.xyz) connected to the open switch.

## Door sensor isn't working
   Make sure all the wires are properly connected, sometimes they get loose. 
   You can debug by connecting the Arduino to your computer, open Arduino on your computer (install it if you don't have it), open the Serial Monitor (under Tools I think) and typing something. The two bytes sent back and forth will appear. You will need to multiply the first one by 256 and add the second one to that number to get the sensor value. Make sure the door is closed when yo do this. Then, put this value + 50 into main.go to replace X in doorMovingAverage.avg() > X.
   
