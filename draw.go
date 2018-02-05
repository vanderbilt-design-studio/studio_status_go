package main

import (
	"image/color"
	"strings"
	"fmt"
	"github.com/sameer/openvg"
	"time"
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
	openvg.Start(s.Width, s.Height)
	openvg.Background(openvg.UnwrapRGB(s.BackgroundFill)) // Fill BG vals
	s.drawDesignStudio()                                  // Draw the words "Design Studio"
	s.drawOpen(s.Open)                                    // Handles whether the studio is open
	s.drawMentorOnDuty()                                  // Mentor name if there is one on duty
	s.drawTime()
	openvg.End()
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


