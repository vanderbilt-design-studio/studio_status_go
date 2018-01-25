package main

import "time"
type mentor_shift struct {
	start    time.Time
	weekday time.Weekday
	duration time.Duration
	name     string
}

type mentor_shifts []mentor_shift

func MustParse(layout, value string) time.Time {
	if t, err := time.Parse(layout, value + " CST"); err != nil {
		panic(err)
	} else {
		return t
	}
}

const mentorTimeLayout = "3:04PM MST"
const mentorDefaultShiftDuration = time.Duration(time.Hour * 2)

// Mentor names array. Each row is a day of the week (sun, mon, ..., sat). Each element in a
// row is a mentor timeslot starting at 12PM, where each slot is 2 hours long.
var shifts = mentor_shifts{
	{MustParse(mentorTimeLayout, "2:00PM"), time.Sunday, mentorDefaultShiftDuration, "Iliya L"},
	{MustParse(mentorTimeLayout, "4:00PM"), time.Sunday, mentorDefaultShiftDuration, "Nick B"},
	{MustParse(mentorTimeLayout, "6:00PM"), time.Sunday, mentorDefaultShiftDuration, "Foard N"},
	{MustParse(mentorTimeLayout, "2:00PM"), time.Monday, mentorDefaultShiftDuration, "Lin L"},
	{MustParse(mentorTimeLayout, "4:00PM"), time.Monday, mentorDefaultShiftDuration, "Amaury P"},
	{MustParse(mentorTimeLayout, "6:00PM"), time.Monday, mentorDefaultShiftDuration, "Jeremy D"},
	{MustParse(mentorTimeLayout, "8:00PM"), time.Monday, mentorDefaultShiftDuration, "Kurt L"},
	{MustParse(mentorTimeLayout, "2:00PM"), time.Tuesday, mentorDefaultShiftDuration, "Sophia Z"},
	{MustParse(mentorTimeLayout, "4:00PM"), time.Tuesday, mentorDefaultShiftDuration, "Emily Mk"},
	{MustParse(mentorTimeLayout, "6:00PM"), time.Tuesday, mentorDefaultShiftDuration, "Jonah H"},
	{MustParse(mentorTimeLayout, "2:00PM"), time.Wednesday, mentorDefaultShiftDuration, "Eric N"},
	{MustParse(mentorTimeLayout, "4:00PM"), time.Wednesday, mentorDefaultShiftDuration, "Lauren B"},
	{MustParse(mentorTimeLayout, "6:00PM"), time.Wednesday, mentorDefaultShiftDuration, "Sameer P"},
	{MustParse(mentorTimeLayout, "8:00PM"), time.Wednesday, mentorDefaultShiftDuration, "Christina H"},
	{MustParse(mentorTimeLayout, "2:00PM"), time.Thursday, mentorDefaultShiftDuration, "Alex B"},
	{MustParse(mentorTimeLayout, "4:00PM"), time.Thursday, mentorDefaultShiftDuration, "Emily Mc"},
	{MustParse(mentorTimeLayout, "6:00PM"), time.Thursday, mentorDefaultShiftDuration, "Braden B"},
	{MustParse(mentorTimeLayout, "8:00PM"), time.Thursday, mentorDefaultShiftDuration, "Jill B"},
	{MustParse(mentorTimeLayout, "2:00PM"), time.Friday, mentorDefaultShiftDuration, "Dominic G"},
	{MustParse(mentorTimeLayout, "4:00PM"), time.Friday, mentorDefaultShiftDuration, "Josh P"},

	{MustParse(mentorTimeLayout, "3:00PM"), time.Tuesday, mentorDefaultShiftDuration, "David L"},
	{MustParse(mentorTimeLayout, "5:00PM"), time.Tuesday, mentorDefaultShiftDuration, "Yunyu L"},
	{MustParse(mentorTimeLayout, "5:00PM"), time.Wednesday, mentorDefaultShiftDuration, "Swapnil P"},
	{MustParse(mentorTimeLayout, "3:00PM"), time.Thursday, mentorDefaultShiftDuration, "Joey H"},
	{MustParse(mentorTimeLayout, "5:00PM"), time.Thursday, mentorDefaultShiftDuration, "Jesse L"},
}

func (this mentor_shifts) getMentorsOnDuty() (mentorsOnDuty []mentor_shift) {
	now := time.Now()
	y, m, d := now.Date()
	mentorsOnDuty = make([]mentor_shift, 0, 2)
	for _, shift := range this {
		shiftStart := shift.start.AddDate(y, int(m) - 1, d - 1)
		if shift.weekday == now.Weekday() && shiftStart.Before(now) && shiftStart.Add(shift.duration).After(now) {
			mentorsOnDuty = append(mentorsOnDuty, shift)
		}
	}
	return
}
