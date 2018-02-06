package main

import (
	"fmt"

	"bytes"
	"github.com/sameer/fsm/moore"
	"github.com/vanderbilt-design-studio/studio-statistics"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var sigstate atomic.Value

func spawnSignalBroadcaster() {
	sigchan := make(chan os.Signal, 2)
	signal.Notify(sigchan, os.Interrupt, os.Kill)
	sigstate.Store("")
	go func() {
		v := <-sigchan
		sigstate.Store(v.String())
	}()
}

const postUrl = "https://ds-sign.yunyul.in"
const logFilename = "activity.log"

var logMutex sync.Mutex

func spawnLogAndPost() chan SignState {
	const logAndPostPeriod = time.Duration(5 * time.Second)
	c := make(chan SignState)
	go func(stateChannel chan SignState) {
		tick := time.NewTicker(logAndPostPeriod)
		logFile, err := os.OpenFile(logFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		shouldLog := true
		if err != nil {
			fmt.Println(err)
			shouldLog = false
		} else {
			defer logFile.Close()
		}
		for sigstate.Load() == "" {
			state := <-stateChannel
			select {
			case <-tick.C:
				state.Post()
				if shouldLog {
					logMutex.Lock()
					state.Log(logFile)
					logMutex.Unlock()
				}
			default:
				continue
			}
		}
	}(c)
	return c
}

func spawnStatsPoster() {
	const statsPostPeriod = time.Duration(1 * time.Minute)
	go func() {
		tick := time.NewTicker(statsPostPeriod)
		for sigstate.Load() == "" {
			select {
			case <-tick.C:
				x_api_key := os.Getenv("x_api_key")
				if x_api_key == "" {
					continue
				}
				logMutex.Lock()
				content, err := ioutil.ReadFile(logFilename)
				logMutex.Unlock()
				if err != nil {
					continue
				}
				pr, pw := io.Pipe()
				req, err := http.NewRequest("POST", "http://spuri.io/studio-statistics.png", pr)
				if err != nil {
					fmt.Println("Failed to prepare post request:", err)
					pr.Close()
					pw.Close()
					continue
				}
				req.Header.Add("content-type", "image/png")
				req.Header.Add("x-api-key", x_api_key)
				go func() {
					_, err = http.DefaultClient.Do(req)
					pr.Close()
				}()
				if err := studio_statistics.MakeGraph(bytes.NewReader(content), pw); err != nil {
					fmt.Println("Errored in trying to make graph", err)
				}
				pw.Close()
			}
		}
	}()
}

func (s *SignState) Post() {
	x_api_key := os.Getenv("x_api_key")
	if x_api_key == "" {
		return
	}

	// TODO: grab mentor on duty
	if s.Subtitle != "" {
		s.Subtitle = makeMentorOnDutyStr(s.Subtitle, false) + s.Subtitle
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
	req.Header.Add("x-api-key", x_api_key)

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
	s := state.(*SignState)
	s.draw() // Do draw commands
	s.LogAndPostChan <- *s
	s.DoRelay()
}
