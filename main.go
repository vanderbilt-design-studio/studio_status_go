package main

import (
	"github.com/sameer/openvg"
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
	tick, _ = time.ParseDuration("32ms")
	BLUE = openvg.RGB{0, 67, 123}
	GREEN  = openvg.RGB{0, 95, 77}
	PURPLE = openvg.RGB{157, 0, 113}
	BLACK  = openvg.RGB{0, 0, 0}
	BROWN  = openvg.RGB{98, 51, 30}
	RED    = openvg.RGB{199, 0, 43}
	ORANGE = openvg.RGB{255, 104, 2}
	YELLOW = openvg.RGB{255, 178, 0}
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
	{"", "Juliana Soltys", "Jonah H.", "", "Jillian B"},
	{"", "Eric N", "Lin Liu", "Sameer P", "Alex B"},
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
	serialConf := &serial.Config{Name: "/dev/ttyACM0", Baud: 600, ReadTimeout: time.Millisecond * 700}
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
	if logo != nil {
		openvg.Img(0, 0, logo)
	}
	openvg.FillRGB(BLACK.Red, BLACK.Green, BLACK.Blue, 1)
	openvg.TextMid(960, 1080 - openvg.TextHeight(defaultFont, 100) - openvg.TextDepth(defaultFont, 100), "Design Studio", defaultFont, 100)
}

func isOpen() bool {
	dayOfWeek := dow(time.Now().Day(), int(time.Now().Month()), time.Now().Year())
	currentHour := time.Now().Hour()
	isOpen := false
	if currentHour >= 12 && dayOfWeek > -1 && dayOfWeek < 7 {
		endOfDayHour := len(names[dayOfWeek])*2 + 12
		idx := (currentHour - 12) / 2
		if currentHour < endOfDayHour && len(names[dayOfWeek]) > idx && idx > -1 && len(names[dayOfWeek][idx]) != 0 {
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
	openvg.FillRGB(fill.Red, fill.Green, fill.Blue, 1)
	openvg.TextMid(960, openvg.TextDepth(defaultFont, 300) + openvg.TextHeight(defaultFont, 100) + openvg.TextHeight(defaultFont, 100), text, defaultFont, 300)
}

func drawMentorOnDuty() {
	if isOpen() && getSwitchValue() == openNormal {
		openvg.FillRGB(BLACK.Red, BLACK.Green, BLACK.Blue, 1)
		dutyStr := "Mentor on Duty: "
		now := time.Now()
		dutyStr += names[dow(now.Day(), int(now.Month()), now.Year())][((now.Hour() - 12) / 2)]
		openvg.TextMid(960, openvg.TextDepth(defaultFont, 100), dutyStr, defaultFont, 100)
	}
}

const (
	openNormal   = 1
	openForced   = 2
	closedForced = 0
	defaultFont = "mono"
)

func getSwitchValue() int {
	if isGPIOAvailable {
		openOne, err := hwio.DigitalRead(gpio17)
		if err == nil {
			if openOne == hwio.HIGH {
				return openNormal
			} else {
				openTwo, err := hwio.DigitalRead(gpio27)
				if err == nil {
					if openTwo == hwio.HIGH {
						return openForced
					} else {
						return closedForced
					}
				} else {
					fmt.Println("gpio27 err ", err)
				}
			}
		} else {
			fmt.Println("gpio17 err ", err)
		}
	}
	return openNormal
}
