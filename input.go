package main

import (
	"fmt"

	"github.com/mrmorphic/hwio"
	"github.com/sameer/fsm/moore"
)

type SignInput struct {
	gpio17, gpio27 hwio.Pin // BCM Pin 17, 27 (https://pinout.xyz/)
	gpio26         hwio.Pin
}

func (si *SignInput) init() {
	var err error

	si.gpio17, err = hwio.GetPin("gpio17")
	if si.gpio17 == 0 {
		fmt.Println("gpio17 ", err)
	} else {
		hwio.PinMode(si.gpio17, hwio.INPUT)
	}

	si.gpio27, err = hwio.GetPin("gpio27")
	if si.gpio27 == 0 {
		fmt.Println("gpio27 ", err)
	} else {
		hwio.PinMode(si.gpio27, hwio.INPUT)
	}

	si.gpio26, err = hwio.GetPin("gpio26")
	if si.gpio26 == 0 {
		fmt.Println("gpio26 ", err)
	} else {
		hwio.PinMode(si.gpio26, hwio.INPUT)
	}
}

func (si *SignInput) finish() {
	if si.gpio17 != 0 {
		hwio.ClosePin(si.gpio17)
	}
	if si.gpio27 != 0 {
		hwio.ClosePin(si.gpio27)
	}
	if si.gpio26 != 0 {
		hwio.ClosePin(si.gpio26)
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
	if si.gpio17 != 0 && si.gpio27 != 0 {
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

func (si *SignInput) IsDoorOpen() bool {
	if si.gpio26 == 0 {
		return true
	}

	if result, err := hwio.DigitalRead(si.gpio26); err != nil {
		return true
	} else {
		// If the door is open, sensor will output LOW, else HIGH.
		// If the sensor is disconnected, defaults to LOW which returns the default value.
		return result == hwio.LOW
	}
}

func (si *SignInput) IsThereMotion() bool {
	// Sensor not installed
	return false
}

var inputState *SignInput

var inputFunction = func() moore.InputFunction {
	inputState = &SignInput{}
	inputState.init()
	return func() moore.Input {
		return inputState
	}
}()
