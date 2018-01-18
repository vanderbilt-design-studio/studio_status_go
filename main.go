package main

import (
	"github.com/sameer/openvg"
	"github.com/sameer/fsm/moore"
	"time"
	"image/color"
	"github.com/tarm/serial"
)

const tick = time.Duration(1000 / 30 * time.Millisecond) // convert TPS to useful number
const notifyPeriod = time.Duration(5 * time.Second)

// Mentor names array. Each row is a day of the week (sun, mon, ..., sat). Each element in a
// row is a mentor timeslot starting at 12PM, where each slot is 2 hours long.
var names = [][]string{
	{"", "", "Nick B", "Foard N", ""},
	{"", "Lin L", "Amaury P", "Jeremy D", "Kurt L"},
	{"", "Sophia Z", "Emily Mk", "Jonah H", ""},
	{"", "Eric N", "Lauren B", "Sameer P", "Christina H"},
	{"", "Alex B", "Emily Mc", "Braden B", "Jill B"},
	{"", "Dominic G"},
	{}} // No Saturday shifts

type SignState struct {
	Init           bool
	Width, Height  int        // Display size
	BackgroundFill color.RGBA // Background fill
	Open           bool
	SwitchValue    SwitchState
	Motion         bool
	Title          string
	Subtitle       string
	NotifyTicker   *time.Ticker
	relayArduino   *serial.Port
}

var transitionFunction moore.TransitionFunction = func(state moore.State, input moore.Input) (moore.State, error) {
	var err error = nil
	s := state.(*SignState)
	i := input.(*SignInput)
	if !s.Init {
		// Init to default state
		s.Width, s.Height = openvg.Init()
		s.BackgroundFill = white
		s.Open = false
		s.SwitchValue = stateClosedForced
		s.Motion = false
		s.Title = "Closed"
		s.Subtitle = ""
		s.NotifyTicker = time.NewTicker(notifyPeriod)
		s.relayArduino = AcquireArduinoUID(32)

		s.Init = true // Mark as succeeded
	}

	// Put inputs into state struct
	s.Open = i.IsOpen()
	s.SwitchValue = i.GetSwitchValue()
	s.Motion = i.IsThereMotion()

	// State-based handling of tile
	if s.Open {
		s.Title = "Open"
	} else {
		s.Title = "Closed"
	}

	// State-based handling of subtitle
	if s.Open && s.SwitchValue == stateOpenNormal {
		now := time.Now()
		// This should never ever fail, because it should've already been checked in isOpen().
		s.Subtitle = "Mentor on Duty: " + names[int(now.Weekday())][((now.Hour() - 12) / 2)]
	} else {
		// Reset output string
		s.Subtitle = ""
	}

	if s == nil { // This is the quit state. Cleanup after ourselves.
		openvg.Finish()
	}
	return s, err
}

func main() {
	mm := moore.Make(
		&SignState{},
		nil,
		transitionFunction,
		inputFunction,
		outputFunction,
	)
	mm.Run(time.NewTicker(tick))
}
