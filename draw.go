package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"strings"
	"time"
)

const defaultFont = "helvetica" // Helvetica font is beautiful for long distance reading.

var (
	// A bunch of standard colors from the official guidelines (somewhere) for road signs
	// to ensure AAA accessiblity. (i.e. red is not pure red so protanomaly colorblind see it)
	blue   = sdl.Color{0, 67, 123, 255}
	green  = sdl.Color{0, 95, 77, 255}
	purple = sdl.Color{157, 0, 113, 255}
	black  = sdl.Color{0, 0, 0, 255}
	brown  = sdl.Color{98, 51, 30, 255}
	red    = sdl.Color{199, 0, 43, 255}
	orange = sdl.Color{255, 104, 2, 255}
	yellow = sdl.Color{255, 178, 0, 255}
	white  = sdl.Color{255, 255, 255, 255}
)
var fullscreen = &sdl.Rect{0, 0, width, height}

func (s *SignState) draw() {
	s.Surface.FillRect(fullscreen, colorToUint32(s.BackgroundFill)) // Fill BG vals
	s.blitDesignStudio()                                            // Draw the words "Design Studio"
	s.blitWhetherOpen(s.Open)                                       // Handles whether the studio is open
	s.blitMentorOnDuty()                                            // Mentor name if there is one on duty
	s.blitTime()
	s.Window.UpdateSurface()
}

var desiredFontSizes = []int {120, 250, 580}
const (
	studioSize       = 250
	titleSize        = 580
	subtitleSize     = 120
	timeSize         = 120
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

func (s *SignState) blitDesignStudio() {
	// Set the drawing color to be white
	str := "Design Studio"
	s.blitGeneric(studioSize, str, white, width/2, int32(s.Fonts[studioSize].Height()/2))
}

func (s *SignState) blitGeneric(size int, text string, color sdl.Color, x, y int32) {
	surf, err := s.Fonts[size].RenderUTF8Blended(text, color)
	if err == nil {
		sw, sh, err := s.Fonts[size].SizeUTF8(text)
		if err == nil {
			surf.Blit(nil, s.Surface, &sdl.Rect{x - int32(sw)/2, y - int32(sh)/2, 0, 0})
		}
	} else {
		fmt.Println(err)
	}
	if surf != nil {
		surf.Free()
	}
}

func (s *SignState) blitWhetherOpen(open bool) {
	// White "Closed" on red background.
	s.BackgroundFill = red
	// White "Open" on green background.
	if open {
		s.BackgroundFill = green
	}
	// Draw that, centered and big.
	s.blitGeneric(titleSize, s.Title, white, width/2, height/2)
}

func (s *SignState) blitMentorOnDuty() {
	// Open + normal operation.
	if s.Open && s.SwitchValue == stateOpenNormal {
		// White text
		s.blitGeneric(subtitleSize, makeMentorOnDutyStr(s.Subtitle, true), white, width*1/8, int32(height*7/8 - s.Fonts[subtitleSize].Height()))
		s.blitGeneric(subtitleSize, s.Subtitle, white, width*1/8, height*7/8)
	}
}

func (s *SignState) blitTime() {
	now := time.Now()
	s.blitGeneric(timeSize, now.Format(time.Kitchen), white, width*7/8, height*7/8)
}
