package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/mrmorphic/hwio"
	"github.com/sameer/fsm/moore"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const tick = time.Duration(1000 / 22 * time.Millisecond) // convert ticks per second to useful number

type SignState struct {
	Init           bool
	BackgroundFill sdl.Color // Background fill
	Window         *sdl.Window
	Renderer       *sdl.Renderer
	Fonts          map[int]*ttf.Font
	Open           bool
	DoorOpen       bool
	SwitchValue    SwitchState
	Motion         bool
	Title          string
	Subtitle       string
	LogAndPostChan chan SignState

	gpio22 hwio.Pin
}

func initState(s *SignState) (*SignState, error) {
	// Init to default state
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	} else {
		if i, err := sdl.ShowCursor(sdl.QUERY); i == sdl.ENABLE && err == nil {
			sdl.ShowCursor(sdl.DISABLE)
			if i, err := sdl.ShowCursor(sdl.QUERY); i == sdl.ENABLE && err != nil {
				fmt.Println("failed to hide cursor")
			}
		}
	}
	if window, rend, err := sdl.CreateWindowAndRenderer(width, height, sdl.WINDOW_FULLSCREEN|sdl.WINDOW_SHOWN|sdl.WINDOW_BORDERLESS); err != nil {
		return nil, err
	} else {
		s.Window, s.Renderer = window, rend
	}
	if err := ttf.Init(); err != nil {
		return nil, err
	} else {
		s.Fonts = make(map[int]*ttf.Font)
		for _, size := range desiredFontSizes {
			if font, err := ttf.OpenFont(font, size); err != nil {
				return nil, err
			} else {
				font.SetStyle(ttf.STYLE_BOLD)
				font.SetHinting(ttf.HINTING_MONO)
				s.Fonts[size] = font
			}
		}
	}

	s.BackgroundFill = white
	s.Open = false
	s.DoorOpen = false
	s.SwitchValue = stateClosedForced
	s.Motion = false
	s.Title = "Closed"
	s.Subtitle = ""
	spawnSignalBroadcaster()
	spawnSDLEventWaiter()
	s.LogAndPostChan = spawnLogAndPost()
	spawnStatsPoster()
	var err error
	if s.gpio22, err = hwio.GetPin("gpio22"); err != nil {
		hwio.PinMode(s.gpio22, hwio.OUTPUT)
	}

	s.Init = true // Mark as succeeded
	return s, nil
}

var transitionFunction moore.TransitionFunction = func(state moore.State, input moore.Input) (moore.State, error) {
	s := state.(*SignState)
	i := input.(*SignInput)
	if !s.Init {
		if _, err := initState(s); err != nil {
			return nil, err
		}
	}

	// Put inputs into state struct
	s.Open, s.DoorOpen = i.IsOpen()
	s.SwitchValue = i.GetSwitchValue()
	s.Motion = i.IsThereMotion()

	// State-based handling of tile
	if s.Open {
		s.Title = "Open"
	} else {
		s.Title = "Closed"
	}

	// State-based handling of subtitle
	if s.Open && s.SwitchValue == stateShifts {
		s.Subtitle = ""
		for _, mentorShift := range shifts.getMentorsOnDuty() {
			if s.Subtitle != "" {
				s.Subtitle += " & "
			}
			s.Subtitle += mentorShift.name
		}
	} else if !s.Open && s.SwitchValue == stateShifts { // Show when the studio opens next if there are shifts today
		nextShifts := shifts.getNextMentorsOnDutyToday()
		if len(shifts.getMentorsOnDuty()) > 0 || len(nextShifts) == 0 {
			// TODO: How to handle a missed shift in between other shifts?
			// If there is supposed to be a shift right now and it's closed, we know that the opens at time is probably
			// wrong so we shouldn't misinform the users. What about a shift that is separated from other shifts? Should
			// we still say anything if that shift was missed? i.e. the possibility that there is a day where no one is
			// on duty, due to a school holiday or other reason. For now, we depend upon a mentor to switch the sign to
			// force closed to indicate that we shouldn't tell anyone when it opens.
			s.Subtitle = "?"
		} else if len(nextShifts) > 0 {
			s.Subtitle = nextShifts[0].time(time.Now().Date()).Format(time.Kitchen)
		} else {
			s.Subtitle = ""
		}
	} else {
		s.Subtitle = ""
	}

	if reason := signalStateStr.Load(); reason != "" {
		fmt.Print("Gracefully shutting down because of \"", reason, "\"...")
		if s.gpio22 != 0 {
			hwio.ClosePin(s.gpio22)
		}
		s.Window.Destroy()
		for _, font := range s.Fonts {
			font.Close()
		}
		s = nil
	}

	if s == nil { // This is the quit state. Cleanup after ourselves.
		inputState.finish()
		ttf.Quit()
		sdl.Quit()
		fmt.Println("done!")
		return nil, nil
	}
	return s, nil
}

func main() {
	mm := moore.Make(
		&SignState{},
		nil,
		transitionFunction,
		inputFunction,
		outputFunction,
	)
	if os.Getenv("DEV") != "" {
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	} else {
		go func() {
			http.ListenAndServe("0.0.0.0:6060", nil)
		}()
	}
	mm.Run(time.NewTicker(tick))
}
