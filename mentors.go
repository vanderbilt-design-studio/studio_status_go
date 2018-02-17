package main

import "time"

type mentorShift struct {
	hour     int
	weekday  time.Weekday
	duration time.Duration
	name     string
}

type mentorShifts [25]mentorShift

const mentorDefaultShiftDuration = time.Duration(time.Hour * 2)

// Mentor names array. Each row is a day of the week (sun, mon, ..., sat). Each element in a
// row is a mentor timeslot starting at 12PM, where each slot is 2 hours long.
var shifts = mentorShifts{
	{14, time.Sunday, mentorDefaultShiftDuration, "Iliya L"},
	{16, time.Sunday, mentorDefaultShiftDuration, "Nick B"},
	{18, time.Sunday, mentorDefaultShiftDuration, "Foard N"},
	{14, time.Monday, mentorDefaultShiftDuration, "Lin L"},
	{16, time.Monday, mentorDefaultShiftDuration, "Amaury P"},
	{18, time.Monday, mentorDefaultShiftDuration, "Jeremy D"},
	{20, time.Monday, mentorDefaultShiftDuration, "Kurt L"},
	{14, time.Tuesday, mentorDefaultShiftDuration, "Sophia Z"},
	{16, time.Tuesday, mentorDefaultShiftDuration, "Emily Mk"},
	{18, time.Tuesday, mentorDefaultShiftDuration, "Christina H"},
	{14, time.Wednesday, mentorDefaultShiftDuration, "Eric N"},
	{16, time.Wednesday, mentorDefaultShiftDuration, "Lauren B"},
	{18, time.Wednesday, mentorDefaultShiftDuration, "Sameer P"},
	{14, time.Thursday, mentorDefaultShiftDuration, "Alex B"},
	{16, time.Thursday, mentorDefaultShiftDuration, "Emily Mc"},
	{18, time.Thursday, mentorDefaultShiftDuration, "Braden B"},
	{20, time.Thursday, mentorDefaultShiftDuration, "Jill B"},
	{14, time.Friday, mentorDefaultShiftDuration, "Dominic G"},
	{16, time.Friday, mentorDefaultShiftDuration, "Josh P"},

	{15, time.Monday, mentorDefaultShiftDuration, "Sabina S"},
	{15, time.Tuesday, mentorDefaultShiftDuration, "David L"},
	{17, time.Tuesday, mentorDefaultShiftDuration, "Yunyu L"},
	{17, time.Wednesday, mentorDefaultShiftDuration, "Swapnil P"},
	{15, time.Thursday, mentorDefaultShiftDuration, "Joey H"},
	{17, time.Thursday, mentorDefaultShiftDuration, "Jesse L"},
}

func (this mentorShifts) getMentorsOnDuty() (mentorsOnDuty []mentorShift) {
	now := time.Now()
	y, m, d := now.Date()
	for _, shift := range this {
		shiftStart := time.Date(y, m, d, shift.hour, 0, 0, 0, time.Local)
		if shift.weekday == now.Weekday() && shiftStart.Before(now) && shiftStart.Add(shift.duration).After(now) {
			mentorsOnDuty = append(mentorsOnDuty, shift)
		}
	}
	return
}
