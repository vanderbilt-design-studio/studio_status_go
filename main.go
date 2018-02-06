package main

import (
	"fmt"
	"github.com/sameer/fsm/moore"
	"github.com/sameer/openvg"
	"github.com/tarm/serial"
	"image/color"
	"time"
)

const tick = time.Duration(1000 / 30 * time.Millisecond) // convert ticks per second to useful number

type SignState struct {
	Init           bool
	Width, Height  int        // Display size
	BackgroundFill color.RGBA // Background fill
	Open           bool
	SwitchValue    SwitchState
	Motion         bool
	Title          string
	Subtitle       string
	LogAndPostChan chan SignState

	relayArduino *serial.Port
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
		spawnSignalBroadcaster()
		s.LogAndPostChan = spawnLogAndPost()
		spawnStatsPoster()
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
		s.Subtitle = ""
		for _, mentorShift := range shifts.getMentorsOnDuty() {
			if s.Subtitle != "" {
				s.Subtitle += " & "
			}
			s.Subtitle += mentorShift.name
		}
	} else {
		// Reset output string
		s.Subtitle = ""
	}

	if sigstate.Load() != "" {
		s = nil
	}

	if s == nil { // This is the quit state. Cleanup after ourselves.
		inputState.finish()
		openvg.Finish()
		if s.relayArduino != nil {
			s.relayArduino.Flush()
			s.relayArduino.Close()
		}
		return nil, nil
	} else {
		return s, err
	}
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
