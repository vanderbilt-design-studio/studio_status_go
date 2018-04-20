package main

import (
	"bytes"
	"fmt"
	"github.com/sameer/fsm/moore"
	"github.com/vanderbilt-design-studio/studio-statistics"
	"github.com/veandco/go-sdl2/sdl"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

var signalStateStr atomic.Value

func spawnSignalBroadcaster() {
	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, os.Kill, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signalStateStr.Store("")
	go func() {
		signalListenTicker := time.NewTicker(time.Millisecond * 50)
		for signalStateStr.Load() == "" || signalStateStr.Load() == nil {
			<-signalListenTicker.C
			select {
			case v := <-signalChan:
				signalStateStr.Store(v.String())
			default:
				continue
			}
		}
	}()
}

func spawnSDLEventWaiter() {
	go func() {
		for signalStateStr.Load() == "" || signalStateStr.Load() == nil {
			event := sdl.WaitEventTimeout(50)
			switch event.(type) {
			case *sdl.QuitEvent:
				signalStateStr.Store("SDL quit event issued")
			case *sdl.KeyboardEvent:
				if ke := event.(*sdl.KeyboardEvent); ke.Keysym.Sym == sdl.K_ESCAPE || ke.Keysym.Sym == sdl.K_q {
					signalStateStr.Store("SDL keypress quit event issued")
				}
			}
		}
	}()
}

const postUrl = "https://ds-sign.yunyul.in"
const logFilename = "activity.log"

func spawnLogAndPost() chan SignState {
	const logAndPostPeriod = time.Duration(1 * time.Second)
	c := make(chan SignState)
	go func(stateChannel chan SignState) {
		isDev := os.Getenv("DEV") != ""
		tick := time.NewTicker(logAndPostPeriod)
		logFile, err := os.OpenFile(logFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		shouldLog := !isDev
		if err != nil {
			fmt.Println(err)
			shouldLog = false
		} else {
			defer logFile.Close()
		}
		for signalStateStr.Load() == "" || signalStateStr.Load() == nil {
			state := <-stateChannel
			select {
			case <-tick.C:
				if !isDev {
					state.Post()
				}
				if shouldLog {
					state.Log(logFile)
				}
			default:
				continue
			}
		}
	}(c)
	return c
}

func spawnStatsPoster() {
	const statsPostPeriod = time.Duration(2 * time.Minute)
	go func() {
		tick := time.NewTicker(statsPostPeriod)
		for range tick.C {
			fmt.Println("Beginning post...")
			xApiKey := os.Getenv("x_api_key")
			if xApiKey == "" {
				fmt.Println("No api key, continuing")
				continue
			}
			if os.Getenv("DEV") != "" {
				fmt.Println("Dev env, continuing")
				continue
			}
			fmt.Println("Reading file")
			content, err := ioutil.ReadFile(logFilename)
			if err != nil {
				fmt.Println("Error in accessing file:", err)
				continue
			}
			var buf bytes.Buffer
			fmt.Println("Making graph")
			if err := studio_statistics.MakeGraph(bytes.NewReader(content), &buf); err != nil {
				fmt.Println("Error in trying to make graph", err)
			}
			req, err := http.NewRequest("POST", "http://spuri.io/studio-statistics.png", &buf)
			if err != nil {
				fmt.Println("Failed to prepare post request:", err)
				continue
			}
			req.Header.Add("content-type", "image/png")
			req.Header.Add("content-length", strconv.Itoa(buf.Len()))
			req.Header.Add("x-api-key", xApiKey)
			if _, err := http.DefaultClient.Do(req); err != nil {
				fmt.Println("Error in trying to post data", err)
			}
			fmt.Println("Stats posted with size", buf.Len())
		}
	}()
}

func (s *SignState) Post() {
	xApiKey := os.Getenv("x_api_key")
	if xApiKey == "" {
		return
	}

	// TODO: grab mentor on duty
	if s.Subtitle != "" { // There is a subtitle
		if !s.Open && s.SwitchValue == stateShifts { // whetherOpens text
			if s.Subtitle == "?" { // Unknown studio dynamics from whetherOpen
				s.Subtitle = ""
			} else if _, err := time.Parse(s.Subtitle, time.Kitchen); err != nil { // Put opens at time
				s.Subtitle = whetherOpensOpenAt + s.Subtitle
			} else { // This should never be the case, but who knows -- might as well be safe
				s.Subtitle = ""
			}
		} else {
			s.Subtitle = makePluralHandlingMentorString(s.Subtitle, false) + s.Subtitle
		}
	} else if !s.Open && s.SwitchValue == stateShifts {
		s.Subtitle = whetherOpensNotOpen
	}
	payload := strings.NewReader(fmt.Sprintf(`{"bgColor": "rgb(%v,%v,%v)", "title": "%v", "subtitle": "%v"}`,
		s.BackgroundFill.R, s.BackgroundFill.G, s.BackgroundFill.B,
		s.Title,
		s.Subtitle,
	))

	req, err := http.NewRequest("POST", postUrl, payload)
	if err != nil {
		fmt.Println("Failed to prepare post request:", err)
		return
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-api-key", xApiKey)

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Failed to post data:", err)
	}
}

func (s *SignState) Log(w io.Writer) error {
	csvLine := fmt.Sprintf("%v,%v,%v,%v\n", time.Now().Format(time.RFC3339Nano), s.Open, s.SwitchValue, s.Motion)
	if _, err := w.Write([]byte(csvLine)); err != nil {
		return err
	}
	return nil
}

func (s *SignState) DoRelay() {
	if s.relayArduino != nil {
		if s.Open {
			s.relayArduino.Write([]byte{relayChange, 2})
		} else {
			s.relayArduino.Write([]byte{relayChange, 0})
		}
	}
}

var outputFunction moore.OutputFunction = func(state moore.State) {
	if state == nil {
		return
	}
	s := state.(*SignState)
	s.draw() // Do draw commands
	s.LogAndPostChan <- *s
	s.DoRelay()
}
