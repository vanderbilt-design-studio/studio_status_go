package main

import (
	"bufio"
	"github.com/ajstarks/openvg"
	"github.com/mrmorphic/hwio"
	"github.com/tarm/serial"
	"os"
	"fmt"
	"time"
	"image"
	_ "image/png"
)

var (
	width    int
	height   int
	exitCode = "dsexit"
)

func main() {
	width, height = openvg.Init()
	defer openvg.Finish()
	setup()
	for {
		openvg.Start(width, height)
		draw()
		openvg.End()
	}
}

var names = [][]string{
	{"", "", "Dominic G", "Olivia C", "Foard N"},
	{"", "Juliana Soltys", "Jonah H.", "", "Jillian B"},
	{"", "Eric N", "Lin Liu", "Sameer P", "Alex B"},
	{"", "Lauren B", "Christina H", "Sophia Z", "Taylor P"},
	{"", "Jeremy D", "Illiya", "Emily M", "Nicholas B"},
	{"Liam K", "Josh P"},
	{}}

var (
	BLUE   = [3]uint8{0, 67, 123}
	GREEN  = [3]uint8{0, 95, 77}
	PURPLE = [3]uint8{157, 0, 113}
	BLACK  = [3]uint8{0, 0, 0}
	BROWN  = [3]uint8{98, 51, 30}
	RED    = [3]uint8{199, 0, 43}
	ORANGE = [3]uint8{255, 104, 2}
	YELLOW = [3]uint8{255, 178, 0}
)

var (
	isGPIOAvailable        = false
	isDoorArduinoAvailable = false
	gpio17                 hwio.Pin
	gpio27                 hwio.Pin
	doorArduino            *serial.Port = nil
	logo                   image.Image
)

func setup() {
	if isGPIOAvailable {
		var err error = nil
		gpio17, err = hwio.GetPin("gpio17")
		if err != nil {
			isGPIOAvailable = false
		}
		gpio27, err = hwio.GetPin("gpio27")
		if err != nil {
			isGPIOAvailable = false
		}
		if isGPIOAvailable {
			hwio.PinMode(gpio17, hwio.INPUT)
			hwio.PinMode(gpio27, hwio.INPUT)
		}
	}
	serialConf := &serial.Config{Name: "/dev/ttyACM0", Baud: 600, ReadTimeout: time.Millisecond * 700}
	serialPort, err := serial.OpenPort(serialConf)
	if err == nil {
		isDoorArduinoAvailable = true
		fmt.Println("Acquired serial port!")
		doorArduino = serialPort
	}
	if file, err := os.Open("logo.png"); err != nil {
		logo, _, _ = image.Decode(file)
		file.Close()
	} else {
		file.Close()
	}
}

func draw() {
	openvg.BackgroundColor("white") // white background
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

var buf = make([]byte, 10)

func isDoorOpen() bool {
	if !isDoorArduinoAvailable {
		return true
	}
	doorArduino.Write([]byte{0})
	bytesRead, err := doorArduino.Read(buf)
	if bytesRead == 0 || err != nil {
		return false
	}
	return buf[0] == 1
}

func drawDesignStudio() {
	openvg.Img(0, 0, logo)
	openvg.FillRGB(BLACK[0], BLACK[1], BLACK[2], 1)
	openvg.TextMid(960, 0, "Design Studio", "mono", 200)
}

func isOpen() bool {
	dayOfWeek := dow(time.Now().Day(), int(time.Now().Month()), time.Now().Year())
	currentHour := time.Now().Hour()
	isOpen := false
	if currentHour >= 12 && dayOfWeek > -1 && dayOfWeek < 7 {
		endOfDayHour := len(names[dayOfWeek])*2 + 12
		idx := (currentHour - 12) / 2
		if currentHour < endOfDayHour && len(names[dayOfWeek]) > idx && idx > -1 && len(names[dayOfWeek][idx]) == 0 {
			isOpen = true
		}
	}
	switchValue := getSwitchValue()
	if switchValue == openNormal {
		return isOpen && isDoorOpen()
	} else if switchValue == openForced {
		return isDoorOpen()
	} else {
		return false
	}
}

// d = day in month
// m = month (January = 1 : December = 12)
// y = 4 digit year
// Returns 0 = Sunday .. 6 = Saturday
func dow(d, m, y int) int {
	if m < 3 {
		m += 12
		y--
	}
	return (d + int((float32(m)+1.0)*float32(2.6)) + y + int(y/4) + 6*int(y/100) +
		int(y/400) + 6) %
		7
}

func drawOpen(open bool) {
	fill := RED
	text := "Closed"
	if open {
		fill = GREEN
		text = "Open"
	}
	openvg.FillRGB(fill[0], fill[1], fill[2], 1)
	openvg.TextMid(960, 203, text, "mono", 400)
}

func drawMentorOnDuty() {
	if isOpen() && getSwitchValue() == openNormal {
		openvg.FillRGB(BLACK[0], BLACK[1], BLACK[2], 1)
		dutyStr := "Mentor on Duty: "
		now := time.Now()
		dutyStr += names[dow(now.Day(), int(now.Month()), now.Year())][((now.Hour() - 12) / 2)]
		openvg.TextMid(960, 1075, dutyStr, "mono", 100)
	}
}

const (
	openNormal   = 1
	openForced   = 2
	closedForced = 0
)

func getSwitchValue() int {
	if isGPIOAvailable {
		openOne, err := hwio.DigitalRead(gpio17)
		if err != nil {
			if openOne == hwio.HIGH {
				return openNormal
			} else {
				openTwo, err := hwio.DigitalRead(gpio27)
				if err != nil {
					if openTwo == hwio.HIGH {
						return openForced
					} else {
						return closedForced
					}
				}
			}
		}

	}
	return openNormal
}
