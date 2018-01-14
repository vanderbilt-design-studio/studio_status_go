package main

import (
	"github.com/sameer/openvg"
	"time"
	"image/color"
)


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
	s.drawDesignStudio()                         // Draw the words "Design Studio"
	s.drawOpen(s.Open)                           // Handles whether the studio is open
	s.drawMentorOnDuty()                         // Mentor name if there is one on duty
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
