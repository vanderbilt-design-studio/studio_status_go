package main

import (
	"fmt"
	"github.com/mrmorphic/hwio"
	"github.com/sameer/fsm/moore"
	"github.com/tarm/serial"
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
	si.doorArduino = AcquireArduinoUID(16)
}

func (si *SignInput) finish() {
	if si.isGPIOAvailable {
		hwio.CloseAll()
	}
	if si.doorArduino != nil {
		si.doorArduino.Flush()
		si.doorArduino.Close()
	}
}

type SwitchState int

const (
	stateShifts       SwitchState = iota // 0, stateOpenNormal was confusing so it is now stateShifts
	stateOpenForced                      // 1
	stateClosedForced                    // 2
)

func (si *SignInput) IsOpen() (isOpen, isDoorOpen bool) {
	// Logic to determine if the studio is likely open.
	isOpen = len(shifts.getMentorsOnDuty()) > 0
	isDoorOpen = si.IsDoorOpen()
	// Now check the switch state. This is a DPDT switch with the states I (normal), II (force open), and O (force closed)
	switchValue := si.GetSwitchValue()
	if switchValue == stateShifts {
		// Door open + normally open. If a mentor misses their shift & the door is closed, the sign will say
		// that the studio is closed.
		isOpen = isOpen && isDoorOpen
	} else if switchValue == stateOpenForced {
		// As long as the door is open, forced open will work. The door *must* be open just in case anyone accidentally
		// leaves it in forced open.
		isOpen = isDoorOpen
	} else {
		// Forced closed.
		isOpen = false
	}
	return
}

func (si *SignInput) GetSwitchValue() SwitchState {
	if si.isGPIOAvailable {
		// Is this normal open?
		openOne, err := hwio.DigitalRead(si.gpio17)
		if err == nil {
			if openOne == hwio.HIGH { // It is indeed.
				return stateShifts
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
	return stateShifts
}

const (
	doorSensorReq   = 4
	motionSensorReq = 8
	relayChange     = 16
	identReq        = 32
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
	return buf[0] != 0 // 0 == false aka door closed
}

func (si *SignInput) IsThereMotion() bool {
	var buf = make([]byte, 1)
	if si.doorArduino == nil {
		return false
	}
	si.doorArduino.Write([]byte{motionSensorReq})
	bytesRead, err := si.doorArduino.Read(buf)
	if bytesRead == 0 || err != nil {
		return false
	}
	return buf[0] != 0
}

var inputState *SignInput

var inputFunction = func() moore.InputFunction {
	inputState = &SignInput{}
	inputState.isGPIOAvailable = true
	inputState.init()
	return func() moore.Input {
		return inputState
	}
}()
