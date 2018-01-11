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
	"strconv"
)

const (
	ticksPerSecond int = 30 // Framerate of 30FPS (TPS)
)

var (
	width   int // Display size
	height  int
	tick, _ = time.ParseDuration(strconv.Itoa(1000/ticksPerSecond) + "ms") // convert TPS to useful number
	// A bunch of standard colors from the official guidelines (somewhere) for road signs
	// to ensure AAA accessiblity. (i.e. red is not pure red so protanomaly colorblind see it)
	blue   = openvg.RGB{0, 67, 123}
	green  = openvg.RGB{0, 95, 77}
	purple = openvg.RGB{157, 0, 113}
	black  = openvg.RGB{0, 0, 0}
	brown  = openvg.RGB{98, 51, 30}
	red    = openvg.RGB{199, 0, 43}
	orange = openvg.RGB{255, 104, 2}
	yellow = openvg.RGB{255, 178, 0}
	white  = openvg.RGB{255, 255, 255}
	bgfill openvg.RGB // Background fill
)

func main() {
	width, height = openvg.Init() // Start openvg
	defer openvg.Finish()         // will be run at the very end of main()
	setup()
	for { // Loop keeps the TPS at a certain rate.
		start := time.Now()
		openvg.Start(width, height)       // Allow draw commands
		draw()                            // Do draw commands
		openvg.End()                      // Disallow them
		duration := time.Now().Sub(start) // Check how long it took
		if duration < tick {
			time.Sleep(tick - duration) // Wait a bit if it was faster than the target TPS
		}
	}
}

// Mentor names array. Each row is a day of the week (sun, mon, ..., sat). Each element in a
// row is a mentor timeslot starting at 12PM, where each slot is 2 hours long.
var names = [][]string{
	{"", "", "Nick B", "", ""},
	{"", "Lin L", "Amaury P", "Jeremy D", "Kurt L"},
	{"", "Sophia Z", "Emily Mk", "Jonah H", ""},
	{"", "Eric N", "Lauren B", "Sameer P", "Christina H"},
	{"", "Alex B", "Emily Mc", "Braden B", "Jill B"},
	{"", "Dominic G"},
	{}} // No Saturday shifts

var (
	isGPIOAvailable        = true             // Is the device the sign being run on GPIO-enabled? Yes if raspberry pi
	isDoorArduinoAvailable = false            // Determined during setup()
	gpio17                 hwio.Pin           // BCM Pin 17 (https://pinout.xyz/)
	gpio27                 hwio.Pin           // BCM Pin 27
	doorArduino            *serial.Port = nil // A port for transferring data with the photoresistor sketch
	logo                   image.Image        // Dead code ignore this
)

func setup() {
	bgfill = white // BG white by default
	if isGPIOAvailable { // Grab the pins if available
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
	serialPort, err := serial.OpenPort(serialConf) // Try to open a serial port.
	if err == nil { // Success
		isDoorArduinoAvailable = true
		fmt.Println("Acquired serial port!")
		doorArduino = serialPort
	}
	if file, err := os.Open("./logo.png"); err == nil { // Ignore this
		logo, _, _ = image.Decode(file)
		file.Close()
	} else { // This too
		fmt.Println("Error while loading logo.png ", err)
		file.Close()
		logo = nil
	}
}

func draw() {
	openvg.Background(bgfill.Red, bgfill.Green, bgfill.Blue) // Fill BG vals
	drawDesignStudio()                                       // Draw the words "Design Studio"
	drawOpen(isOpen())                                       // Handles whether the studio is open
	drawMentorOnDuty()                                       // Mentor name if there is one on duty
	flipOpenStripServo()                                     // WIP the servo broke last time I tried this
	// TODO: use an NPN transistor instead of servo
}

var (
	servoOpen   = false // Servo states
	firstSwitch = true
)

func flipOpenStripServo() {
	if isGPIOAvailable {
		// Logic to check if the switch is in the wrong state.
		shouldFlip := false
		if firstSwitch {
			firstSwitch = false
			shouldFlip = true
		} else if isOpen() && !servoOpen { // TODO: self, xor this pls
			shouldFlip = true
		} else if !isOpen() && servoOpen {
			shouldFlip = true
		}
		// Reverse state if the servo is not correct
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

// Vars for tracking photoresistor vals.
var buf = make([]byte, 1)

const (
	doorSensorReq   = 2
	motionSensorReq = 4
)

func isDoorOpen() bool {
	if !isDoorArduinoAvailable {
		return true
	}
	doorArduino.Write([]byte{doorSensorReq}) // Send a request for photoresistor value
	bytesRead, err := doorArduino.Read(buf)
	// If we failed to read a value, assume either something is wrong with the hardware
	// i.e. the Arduino was unplugged. The mentor on duty can override the current value
	// by switching it to open override.
	if bytesRead == 0 || err != nil {
		return true
	}
	return buf[0] == 0; // 0 == false aka door closed
	// Add onto the average
	//doorMovingAverage.Add(float64(val))
	//return doorMovingAverage.Avg() > 200 // The door sees light on average
	// TODO: change this so a sudden 30% difference in the moving average will cause a state transition
	// TODO: remove the magic value of 200 because sometimes the sensor changes irreversably because people
	// mess with the arduino. Instead, we can make an assumption that no one will be in the room at 5/6AM and
	// use the sensor values at those times as "darkness".
}

func isThereMotion() bool {
	if !isDoorArduinoAvailable {
		return false
	}
	doorArduino.Write([]byte{motionSensorReq})
	bytesRead, err := doorArduino.Read(buf)
	if bytesRead == 0 || err != nil {
		return false
	}
	return buf[0] == 0;
}

func drawDesignStudio() {
	if logo != nil { // Ignore this
		openvg.Img(0, 0, logo)
	}
	// Set the drawing color to be white
	openvg.FillRGB(white.Red, white.Green, white.Blue, 1)
	// Draw text at a size of 200 in Helvetica Bold
	size := 200
	openvg.TextMid(960, 1080-openvg.TextHeight(defaultFont, size), "Design Studio", defaultFont, size)
}

func isOpen() bool {
	// Parse time down to important values
	t := time.Now()
	dayOfWeek := int(t.Weekday())
	currentHour := t.Hour()
	// Logic to determine if the studio is likely open.
	isOpen := false
	// The studio can only be normally open after 12PM & if the day of week is valid.
	if currentHour >= 12 && dayOfWeek > -1 && dayOfWeek < 7 {
		// Determine when the day's shifts end.
		endOfDayHour := len(names[dayOfWeek])*2 + 12
		// Turn that into an index into the current names array row
		idx := (currentHour - 12) / 2
		// The time right now should be before the end of the shifts, idx should be within
		// the bounds of the array, and this should not be a day with no shifts.
		if currentHour < endOfDayHour && idx < len(names[dayOfWeek]) && idx > -1 && len(names[dayOfWeek][idx]) != 0 {
			isOpen = true // Normal open state
		}
	}
	// Now check the switch state. This is a DPDT switch with the states I (normal), II (force open), and O (force closed)
	switchValue := getSwitchValue()
	if switchValue == stateOpenNormal {
		// Door open + normally open. If a mentor misses their shift & the door is closed, the sign will say
		// that the studio is closed.
		return isOpen && isDoorOpen()
	} else if switchValue == stateOpenForced {
		// As long as the door is open, forced open will work. The door *must* be open just in case anyone accidentally
		// leaves it in forced open.
		return isDoorOpen()
	} else {
		// Forced closed.
		return false
	}
}

func drawOpen(open bool) {
	// White "Closed" on red background.
	fill := white
	bgfill = red
	text := "Closed"
	// White "Open" on green background.
	if open {
		bgfill = green
		text = "Open"
	}
	// Draw that, centered and big.
	openvg.FillRGB(fill.Red, fill.Green, fill.Blue, 1)
	openvg.TextMid(960, openvg.TextDepth(defaultFont, 400)+openvg.TextHeight(defaultFont, 100)+openvg.TextHeight(defaultFont, 100), text, defaultFont, 400)
}

func drawMentorOnDuty() {
	// Open + normal operation.
	if isOpen() && getSwitchValue() == stateOpenNormal {
		// White text
		openvg.FillRGB(white.Red, white.Green, white.Blue, 1)
		dutyStr := "Mentor on Duty: "
		now := time.Now()
		// This should never ever fail, because it should've already been checked in isOpen().
		dutyStr += names[int(now.Weekday())][((now.Hour() - 12) / 2)]
		openvg.TextMid(960, openvg.TextDepth(defaultFont, 100), dutyStr, defaultFont, 100)
	}
}

const (
	stateOpenNormal   = iota        // 0
	stateOpenForced                 // 1
	stateClosedForced               // 2
	defaultFont       = "helvetica" // Helvetica font is beautiful for long distance reading.
)

func getSwitchValue() int {
	if isGPIOAvailable {
		// Is this normal open?
		openOne, err := hwio.DigitalRead(gpio17)
		if err == nil {
			if openOne == hwio.HIGH { // It is indeed.
				return stateOpenNormal
			} else {
				// Is it actually forced open?
				openTwo, err := hwio.DigitalRead(gpio27)
				if err == nil {
					if openTwo == hwio.HIGH { // It is indeed.
						return stateOpenForced
					} else { // The only other possibility is forced closed.
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
	// This will be returned if there are any errors. There's no way to really recover from fatal errors
	// like these without manual intervention, so it is safest to assume normal operation.
	return stateOpenNormal
}