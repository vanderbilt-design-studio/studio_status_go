package main

import (
	"github.com/sameer/openvg"
	"time"
	"image/color"
	"net/http"
	"fmt"
	"strings"
	"os"
	"github.com/sameer/fsm/moore"
	"github.com/mrmorphic/hwio"
)

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
	// TODO: use an NPN transistor instead of servo
}

func (s *SignState) drawDesignStudio() {
	// Set the drawing color to be white
	openvg.FillRGB(openvg.UnwrapRGBA(white))
	// Draw text at a size of 200 in Helvetica Bold
	size := 200
	openvg.TextMid(960, 1080-openvg.TextHeight(defaultFont, size), "Design Studio", defaultFont, size)
}

func (s *SignState) drawOpen(open bool) {
	// White "Closed" on red background.
	fill := white
	s.BackgroundFill = red
	text := "Closed"
	// White "Open" on green background.
	if open {
		s.BackgroundFill = green
		text = "Open"
	}
	// Draw that, centered and big.
	openvg.FillRGB(openvg.UnwrapRGBA(fill))
	openvg.TextMid(960, openvg.TextDepth(defaultFont, 400)+openvg.TextHeight(defaultFont, 100)+openvg.TextHeight(defaultFont, 100), text, defaultFont, 400)
}

func (s *SignState) drawMentorOnDuty() {
	// Open + normal operation.
	if s.Open && s.SwitchValue == stateOpenNormal {
		// White text
		openvg.FillRGB(openvg.UnwrapRGBA(white))
		dutyStr := "Mentor on Duty: "
		now := time.Now()
		// This should never ever fail, because it should've already been checked in isOpen().
		dutyStr += names[int(now.Weekday())][((now.Hour() - 12) / 2)]
		openvg.TextMid(960, openvg.TextDepth(defaultFont, 100), dutyStr, defaultFont, 100)
	}
}

const postUrl = "https://ds-sign.yunyul.in"
const logFilename = "activity.log"

func (s *SignState) Notify() {
	select {
	case <-s.NotifyTicker.C: // If it is time to do a post!
		s.Post()
		s.Log()
	default:
		return
	}

}

func (s *SignState) Post() {
	x_api_key := os.Getenv("x_api_key")
	if x_api_key == "" {
		return
	}
	title := "Closed"
	if s.Open {
		title = "Open"
	}
	// TODO: grab mentor on duty
	subtitle := ""
	payload := strings.NewReader(fmt.Sprintf(`{"bgColor": "rgb(%v,%v,%v)", "title": "%v", "subtitle": "%v"}`,
		s.BackgroundFill.R, s.BackgroundFill.G, s.BackgroundFill.B,
		title,
		subtitle,
	))

	req, err := http.NewRequest("POST", postUrl, payload)
	if err != nil {
		fmt.Printf("Failed to prepare post request: %v\n", err)
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-api-key", x_api_key)

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to post data: %v\n", err)
	}
}

func (s *SignState) Log() {
	stateCopy := *s
	go func() {
		logFile, err := os.OpenFile(logFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		defer logFile.Close()
		csvLine := fmt.Sprintf("%v,%v,%v,%v\n", time.Now().Format(time.RFC822), stateCopy.Open, stateCopy.SwitchValue, stateCopy.Motion)
		if _, err = logFile.WriteString(csvLine); err != nil {
			panic(err)
		}
	}()
}

func (s *SignState) DoRelay() {
	if s.isRelayAvailable {
		val := hwio.HIGH
		if !s.Open {
			val = hwio.LOW
		}
		hwio.DigitalWrite(s.gpio22, val)
	}
}

var outputFunction moore.OutputFunction = func(state moore.State) {
	s := state.(*SignState)
	openvg.Start(s.Width, s.Height) // Allow draw commands
	s.draw()                        // Do draw commands
	openvg.End()                    // Disallow them
	s.Notify()
	s.DoRelay()
}
