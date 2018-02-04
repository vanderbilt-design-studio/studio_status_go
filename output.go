package main

import (
	"fmt"

	"github.com/sameer/fsm/moore"
	"github.com/sameer/openvg"
	"sync/atomic"
	"os/signal"
	"image/color"
	"strings"
	"time"
	"sync"
	"os"
	"io"
	"net/http"
	"github.com/vanderbilt-design-studio/studio-statistics"
	"bytes"
	"io/ioutil"
)

var sigstate atomic.Value

func spawnSignalBroadcaster() {
	sigchan := make(chan os.Signal, 2)
	signal.Notify(sigchan, os.Interrupt, os.Kill)
	go func() {
		v := <-sigchan
		sigstate.Store(v.String())
	}()
}

const defaultFont = "helvetica" // Helvetica font is beautiful for long distance reading.

var (
	// A bunch of standard colors from the official guidelines (somewhere) for road signs
	// to ensure AAA accessiblity. (i.e. red is not pure red so protanomaly colorblind see it)
	blue   = color.RGBA{0, 67, 123, 255}
	green  = color.RGBA{0, 95, 77, 255}
	purple = color.RGBA{157, 0, 113, 255}
	black  = color.RGBA{0, 0, 0, 255}
	brown  = color.RGBA{98, 51, 30, 255}
	red    = color.RGBA{199, 0, 43, 255}
	orange = color.RGBA{255, 104, 2, 255}
	yellow = color.RGBA{255, 178, 0, 255}
	white  = color.RGBA{255, 255, 255, 255}
)

func (s *SignState) draw() {
	openvg.Background(openvg.UnwrapRGB(s.BackgroundFill)) // Fill BG vals
	s.drawDesignStudio()                                  // Draw the words "Design Studio"
	s.drawOpen(s.Open)                                    // Handles whether the studio is open
	s.drawMentorOnDuty()                                  // Mentor name if there is one on duty
	s.drawTime()
}

const (
	studioSize       = 200
	titleSize        = 400
	subtitleSize     = 100
	timeSize         = 100
	mentorOnDutyStrf = "Mentor%v on Duty: "
	mentorPrefixStrf = "Mentor%v: "
)

func makeMentorOnDutyStr(subtitle string, onDutyText bool) string {
	multi := ""
	if strings.ContainsRune(subtitle, '&') {
		multi = "s"
	}
	if onDutyText {
		return fmt.Sprintf(mentorOnDutyStrf, multi)
	} else {
		return fmt.Sprintf(mentorPrefixStrf, multi)
	}
}

func (s *SignState) drawDesignStudio() {
	// Set the drawing color to be white
	openvg.FillRGB(openvg.UnwrapRGBA(white))
	// Draw text at a size of 200 in Helvetica Bold
	openvg.TextMid(960, 1080-openvg.TextHeight(defaultFont, studioSize)+openvg.TextDepth(defaultFont, studioSize), "Design Studio", defaultFont, studioSize)
}

func (s *SignState) drawOpen(open bool) {
	// White "Closed" on red background.
	fill := white
	s.BackgroundFill = red
	// White "Open" on green background.
	if open {
		s.BackgroundFill = green
	}
	// Draw that, centered and big.
	openvg.FillRGB(openvg.UnwrapRGBA(fill))
	openvg.TextMid(960, 1080-openvg.TextHeight(defaultFont, studioSize)-openvg.TextDepth(defaultFont, studioSize)-openvg.TextDepth(defaultFont, titleSize)-openvg.TextHeight(defaultFont, titleSize)/2, s.Title, defaultFont, titleSize)
}

func (s *SignState) drawMentorOnDuty() {
	// Open + normal operation.
	if s.Open && s.SwitchValue == stateOpenNormal {
		// White text
		openvg.FillRGB(openvg.UnwrapRGBA(white))
		openvg.Text(0, openvg.TextHeight(defaultFont, subtitleSize)+openvg.TextDepth(defaultFont, subtitleSize), makeMentorOnDutyStr(s.Subtitle, true), defaultFont, subtitleSize)
		openvg.Text(0, openvg.TextDepth(defaultFont, subtitleSize), s.Subtitle, defaultFont, subtitleSize)
	}
}

func (s *SignState) drawTime() {
	now := time.Now()
	openvg.TextEnd(1920, openvg.TextHeight(defaultFont, timeSize)+openvg.TextDepth(defaultFont, timeSize), now.Format(time.Kitchen), defaultFont, timeSize)
}

const postUrl = "https://ds-sign.yunyul.in"
const logFilename = "activity.log"

var logMutex sync.Mutex

func spawnLogAndPost() (chan SignState) {
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
				req, err := http.NewRequest("POST", "https://spuri.io/studio-statistics.png", pr)
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
				studio_statistics.MakeGraph(bytes.NewReader(content), pw)
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

func (s *SignState) Log(w io.Writer) (error) {
	csvLine := fmt.Sprintf("%v,%v,%v,%v\n", time.Now().Format(time.RFC822), s.Open, s.SwitchValue, s.Motion)
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
	openvg.Start(s.Width, s.Height) // Allow draw commands
	s.draw()                        // Do draw commands
	openvg.End()                    // Disallow them
	s.LogAndPostChan <- *s
	s.DoRelay()
}
