package main

import (
	"github.com/sameer/fsm/moore"
	"github.com/tarm/serial"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"time"
	"net/http"
	_ "net/http/pprof"
	"os"
)

const tick = time.Duration(1000 / 30 * time.Millisecond) // convert ticks per second to useful number

const width, height = 1920, 1080

const font = "Helvetica-Bold.ttf"

type SignState struct {
	Init           bool
	BackgroundFill sdl.Color // Background fill
	Window         *sdl.Window
	Surface        *sdl.Surface
	Fonts          map[int]*ttf.Font
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
		if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
			return nil, err
		}
		if window, err := sdl.CreateWindow("", 0, 0, width, height, sdl.WINDOW_FULLSCREEN|sdl.WINDOW_SHOWN); err != nil {
			return nil, err
		} else if surf, err := window.GetSurface(); err == nil {
			s.Window, s.Surface = window, surf
		} else {
			return err, nil
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
		if s.relayArduino != nil {
			s.relayArduino.Flush()
			s.relayArduino.Close()
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
	if os.Getenv("DEV") != "" {
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	}
	mm.Run(time.NewTicker(tick))
}
