package main

import (
	"github.com/mrmorphic/hwio"
	"github.com/tarm/serial"
	"fmt"
	"github.com/sameer/fsm/moore"
	"time"
)

type SignInput struct {
	isGPIOAvailable bool
	gpio17, gpio27  hwio.Pin     // BCM Pin 17, 27 (https://pinout.xyz/)
	doorArduino     *serial.Port // A port for transferring data with the photoresistor sketch
}

func (si *SignInput) init() {
	if si.isGPIOAvailable { // Grab the pins if available
		var err error = nil
		si.gpio17, err = hwio.GetPin("gpio17")
		if err != nil {
			fmt.Println("gpio17 ", err)
			si.isGPIOAvailable = false
		}
		si.gpio27, err = hwio.GetPin("gpio27")
		if err != nil {
			fmt.Println("gpio27 ", err)
			si.isGPIOAvailable = false
		}
		if si.isGPIOAvailable {
			hwio.PinMode(si.gpio17, hwio.INPUT)
			hwio.PinMode(si.gpio27, hwio.INPUT)
		}
	}
	serialConf := &serial.Config{Name: "/dev/ttyACM0", Baud: 9600}
	serialPort, err := serial.OpenPort(serialConf) // Try to open a serial port.
	if err == nil { // Success
		fmt.Println("Acquired serial port!")
		si.doorArduino = serialPort
	}
}

type SwitchState int

const (
	stateOpenNormal   SwitchState = iota // 0
	stateOpenForced                      // 1
	stateClosedForced                    // 2
)

func (si *SignInput) IsOpen() bool {
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
	switchValue := si.GetSwitchValue()
	if switchValue == stateOpenNormal {
		// Door open + normally open. If a mentor misses their shift & the door is closed, the sign will say
		// that the studio is closed.
		return isOpen && si.IsDoorOpen()
	} else if switchValue == stateOpenForced {
		// As long as the door is open, forced open will work. The door *must* be open just in case anyone accidentally
		// leaves it in forced open.
		return si.IsDoorOpen()
	} else {
		// Forced closed.
		return false
	}
}

func (si *SignInput) GetSwitchValue() SwitchState {
	if si.isGPIOAvailable {
		// Is this normal open?
		openOne, err := hwio.DigitalRead(si.gpio17)
		if err == nil {
			if openOne == hwio.HIGH { // It is indeed.
				return stateOpenNormal
			} else {
				// Is it actually forced open?
				openTwo, err := hwio.DigitalRead(si.gpio27)
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

const (
	doorSensorReq   = 2
	motionSensorReq = 4
)

func (si *SignInput) IsDoorOpen() bool {
	var buf = make([]byte, 1)
	if si.doorArduino == nil {
		return true
	}
	si.doorArduino.Write([]byte{doorSensorReq}) // Send a request for photoresistor value
	bytesRead, err := si.doorArduino.Read(buf)
	// If we failed to read a value, assume either something is wrong with the hardware
	// i.e. the Arduino was unplugged. The mentor on duty can override the current value
	// by switching it to open override.
	if bytesRead == 0 || err != nil {
		return true
	}
	return buf[0] != 0; // 0 == false aka door closed
	// Add onto the average
	//doorMovingAverage.Add(float64(val))
	//return doorMovingAverage.Avg() > 200 // The door sees light on average
	// TODO: change this so a sudden 30% difference in the moving average will cause a state transition
	// TODO: remove the magic value of 200 because sometimes the sensor changes irreversably because people
	// mess with the arduino. Instead, we can make an assumption that no one will be in the room at 5/6AM and
	// use the sensor values at those times as "darkness".
}

func (si *SignInput) isThereMotion() bool {
	var buf = make([]byte, 1)
	if si.doorArduino == nil {
		return false
	}
	si.doorArduino.Write([]byte{motionSensorReq})
	bytesRead, err := si.doorArduino.Read(buf)
	if bytesRead == 0 || err != nil {
		return false
	}
	return buf[0] != 0;
}

var inputFunction moore.InputFunction = func() moore.InputFunction {
	input := &SignInput{}
	input.init()
	return func() moore.Input {
		return input
	}
}()
