package main

import (
	"github.com/sameer/openvg"
	"github.com/sameer/fsm/moore"
	"time"
	"image/color"
)

const tick = time.Duration(1000 / 30 * time.Millisecond) // convert TPS to useful number
const notifyPeriod = time.Duration(time.Minute)
const defaultFont = "helvetica" // Helvetica font is beautiful for long distance reading.

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
	NotifyTicker   *time.Ticker
}

var transitionFunction moore.TransitionFunction = func(state moore.State, input moore.Input) (moore.State, error) {
	var err error = nil
	s := state.(*SignState)
	i := input.(*SignInput)
	if s == nil {
		openvg.Finish()
	}
	if !s.Init {
		s.Width, s.Height = openvg.Init() // Start openvg
		s.BackgroundFill = white
		s.Open = false
		s.SwitchValue = stateClosedForced
		s.NotifyTicker = time.NewTicker(notifyPeriod)
		s.Init = true
	}

	s.Open = i.IsOpen()
	s.SwitchValue = i.GetSwitchValue()
	s.Motion = i.IsThereMotion()

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
	<-mm.Fork(time.NewTicker(tick))
}
