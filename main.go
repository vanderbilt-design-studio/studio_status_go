package main

import (
	"github.com/sameer/openvg"
	"github.com/mrmorphic/hwio"
	"github.com/tarm/serial"
	"github.com/RobinUS2/golang-moving-average"
	"os"
	"fmt"
	"time"
	"image"
	_ "image/png"
	"strconv"
)

const (
	ticksPerSecond int = 30
)

var (
	width   int
	height  int
	tick, _ = time.ParseDuration(strconv.Itoa(1000/ticksPerSecond) + "ms")
	blue    = openvg.RGB{0, 67, 123}
	green   = openvg.RGB{0, 95, 77}
	purple  = openvg.RGB{157, 0, 113}
	black   = openvg.RGB{0, 0, 0}
	brown   = openvg.RGB{98, 51, 30}
	red     = openvg.RGB{199, 0, 43}
	orange  = openvg.RGB{255, 104, 2}
	yellow  = openvg.RGB{255, 178, 0}
	white   = openvg.RGB{255, 255, 255}
	bgfill  openvg.RGB
)

func main() {
	width, height = openvg.Init()
	defer openvg.Finish()
	setup()
	for {
		start := time.Now()
		openvg.Start(width, height)
		draw()
		openvg.End()
		duration := time.Now().Sub(start)
		if duration < tick {
			time.Sleep(tick - duration)
		}
	}
}

var names = [][]string{
	{"", "", "Dominic G", "Olivia C", "Foard N"},
	{"", "Juliana S", "Jonah H", "", "Jillian B"},
	{"", "Eric N", "Lin L", "Sameer P", "Alex B"},
	{"", "Lauren B", "Christina H", "Sophia Z", "Taylor P"},
	{"", "Jeremy D", "Illiya", "Emily M", "Nicholas B"},
	{"Liam K", "Josh P"},
	{}}

var (
	isGPIOAvailable        = true
	isDoorArduinoAvailable = false
	gpio17                 hwio.Pin
	gpio27                 hwio.Pin
	doorArduino            *serial.Port = nil
	logo                   image.Image
)

func setup() {
	bgfill = white
	if isGPIOAvailable {
		var err error = nil
		gpio17, err = hwio.GetPin("gpio17")
		if err != nil {
			fmt.Println("gpio17 ", err)
			isGPIOAvailable = false
		}
		gpio27, err = hwio.GetPin("gpio27")
		if err != nil {
			fmt.Println("gpio27 ", err)
			isGPIOAvailable = false
		}
		if isGPIOAvailable {
			hwio.PinMode(gpio17, hwio.INPUT)
			hwio.PinMode(gpio27, hwio.INPUT)
		}
	}
	serialConf := &serial.Config{Name: "/dev/ttyACM0", Baud: 9600}
	serialPort, err := serial.OpenPort(serialConf)
	if err == nil {
		isDoorArduinoAvailable = true
		fmt.Println("Acquired serial port!")
		doorArduino = serialPort
	}
	if file, err := os.Open("./logo.png"); err == nil {
		logo, _, _ = image.Decode(file)
		file.Close()
	} else {
		fmt.Println("Error while loading logo.png ", err)
		file.Close()
		logo = nil
	}
}

func draw() {
	openvg.Background(bgfill.Red, bgfill.Green, bgfill.Blue) // background
	drawDesignStudio()
	drawOpen(isOpen())
	drawMentorOnDuty()
	flipOpenStripServo()
}

var (
	servoOpen   = false
	firstSwitch = true
)

func flipOpenStripServo() {
	if isGPIOAvailable {
		shouldFlip := false
		if firstSwitch {
			firstSwitch = false
			shouldFlip = true
		} else if isOpen() && !servoOpen { // TODO: self, xor this pls
			shouldFlip = true
		} else if !isOpen() && servoOpen {
			shouldFlip = true
		}

		if shouldFlip {
			if isOpen() {
				// stripservo.write(140);
				servoOpen = true
			} else {
				// stripservo.write(65);
				servoOpen = false
			}
		}
	}
}

var buf = make([]byte, 1)
var doorMovingAverage = movingaverage.New(ticksPerSecond*2)

func isDoorOpen() bool {
	if !isDoorArduinoAvailable {
		return true
	}
	doorArduino.Write([]byte{1})
	bytesRead, err := doorArduino.Read(buf)
	if bytesRead == 0 || err != nil {
		return false
	}
	val := int(buf[0])*256
	bytesRead, err = doorArduino.Read(buf)
	val += int(buf[0])
	doorMovingAverage.Add(float64(val))
	return doorMovingAverage.Avg() > 200 // The door sees light on average
}

func drawDesignStudio() {
	if logo != nil {
		openvg.Img(0, 0, logo)
	}
	openvg.FillRGB(white.Red, white.Green, white.Blue, 1)
	size := 200
	openvg.TextMid(960, 1080-openvg.TextHeight(defaultFont, size), "Design Studio", defaultFont, size)
}

func isOpen() bool {
	t := time.Now()
	dayOfWeek := int(t.Weekday())
	currentHour := t.Hour()
	isOpen := false
	if currentHour >= 12 && dayOfWeek > -1 && dayOfWeek < 7 {
		endOfDayHour := len(names[dayOfWeek])*2 + 12
		idx := (currentHour - 12) / 2
		if currentHour < endOfDayHour && idx < len(names[dayOfWeek]) && idx > -1 && len(names[dayOfWeek][idx]) != 0 {
			isOpen = true
		}
	}
	switchValue := getSwitchValue()
	if switchValue == stateOpenNormal {
		return isOpen && isDoorOpen()
	} else if switchValue == stateOpenForced {
		return isDoorOpen()
	} else {
		return false
	}
}

func drawOpen(open bool) {
	fill := white
	bgfill = red
	text := "Closed"
	if open {
		bgfill = green
		text = "Open"
	}
	openvg.FillRGB(fill.Red, fill.Green, fill.Blue, 1)
	openvg.TextMid(960, openvg.TextDepth(defaultFont, 400)+openvg.TextHeight(defaultFont, 100)+openvg.TextHeight(defaultFont, 100), text, defaultFont, 400)
}

func drawMentorOnDuty() {
	if isOpen() && getSwitchValue() == stateOpenNormal {
		openvg.FillRGB(white.Red, white.Green, white.Blue, 1)
		dutyStr := "Mentor on Duty: "
		now := time.Now()
		dutyStr += names[int(now.Weekday())][((now.Hour() - 12) / 2)]
		openvg.TextMid(960, openvg.TextDepth(defaultFont, 100), dutyStr, defaultFont, 100)
	}
}

const (
	stateOpenNormal   = iota
	stateOpenForced
	stateClosedForced
	defaultFont       = "helvetica"
)

func getSwitchValue() int {
	if isGPIOAvailable {
		openOne, err := hwio.DigitalRead(gpio17)
		if err == nil {
			if openOne == hwio.HIGH {
				return stateOpenNormal
			} else {
				openTwo, err := hwio.DigitalRead(gpio27)
				if err == nil {
					if openTwo == hwio.HIGH {
						return stateOpenForced
					} else {
						return stateClosedForced
					}
				} else {
					fmt.Println("gpio27 err ", err)
				}
			}
		} else {
			fmt.Println("gpio17 err ", err)
		}
	}
	return stateOpenNormal
}
