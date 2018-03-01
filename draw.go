package main

import (
	"container/list"
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"hash/crc64"
	"strconv"
	"strings"
	"time"
)

const font = "Helvetica-Bold.ttf" // Helvetica font is beautiful for long distance reading.
const width, height = 1920, 1080

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

func (s *SignState) draw() {
	s.Renderer.SetDrawColor(s.BackgroundFill.R, s.BackgroundFill.G, s.BackgroundFill.B, s.BackgroundFill.A)
	s.Renderer.Clear()
	s.blitDesignStudio()      // Draw the words "Design Studio"
	s.blitWhetherOpen(s.Open) // Handles whether the studio is open
	s.blitMentorOnDuty()      // Mentor name if there is one on duty
	s.blitTime()
	s.Renderer.Present()
}

var desiredFontSizes = [3]int{120, 250, 580}

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
	s.blitCentered(studioSize, str, white, width/2, int32(s.Fonts[studioSize].Height()/2))
}

func (s *SignState) blitCentered(size int, text string, color sdl.Color, x, y int32) {
	sw, sh, err := s.Fonts[size].SizeUTF8(text)
	if err != nil {
		fmt.Println(err)
	} else {
		s.blitLeft(size, text, color, x-int32(sw)/2, y-int32(sh)/2)
	}
}

var cacheList = list.New()

type cachedTexture struct {
	tex      *sdl.Texture
	checksum uint64
}

var crc64Table = crc64.MakeTable(crc64.ISO)

func (s *SignState) blitLeft(size int, text string, color sdl.Color, x, y int32) {
	checksum := crc64.Checksum([]byte(text+strconv.Itoa(size)+strconv.Itoa(int(color.Uint32()))), crc64Table)
	var tex *sdl.Texture
	for e := cacheList.Front(); e != cacheList.Back(); e = e.Next() {
		if e.Value.(cachedTexture).checksum == checksum {
			tex = e.Value.(cachedTexture).tex
			cacheList.MoveToFront(e)
			break
		}
	}
	if tex == nil {
		var err error
		var surf *sdl.Surface
		surf, err = s.Fonts[size].RenderUTF8Blended(text, color)
		if err != nil {
			fmt.Println(err)
			if surf != nil {
				surf.Free()
				surf = nil
			}
		} else {
			tex, err = s.Renderer.CreateTextureFromSurface(surf)
			surf.Free()
			if err != nil {
				if tex != nil {
					tex.Destroy()
					tex = nil
				}
			} else {
				if cacheList.Len() > 6 {
					cacheList.Back().Value.(cachedTexture).tex.Destroy()
					cacheList.Remove(cacheList.Back())
				}
				cacheList.PushFront(cachedTexture{tex, checksum})
			}
		}
	}

	if tex != nil {
		s.Renderer.SetDrawColor(white.R, white.G, white.B, white.A)
		_, _, w, h, err := tex.Query()
		if err == nil {
			s.Renderer.Copy(tex, nil, &sdl.Rect{x, y, w, h})
		}

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
	s.blitCentered(titleSize, s.Title, white, width/2, height*7/16)
}

func (s *SignState) blitMentorOnDuty() {
	// Open + normal operation.
	if s.Open && s.SwitchValue == stateOpenNormal {
		// White text
		s.blitLeft(subtitleSize, makeMentorOnDutyStr(s.Subtitle, true), white, width*1/64, int32(height-s.Fonts[subtitleSize].Height()*2))
		s.blitLeft(subtitleSize, s.Subtitle, white, width*1/64, int32(height-s.Fonts[subtitleSize].Height()))
	}
}

func (s *SignState) blitTime() {
	now := time.Now()
	s.blitCentered(timeSize, now.Format(time.Kitchen), white, width*13/16, height*15/16)
}
