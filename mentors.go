package main

import "time"

type mentorShift struct {
	hour     int
	weekday  time.Weekday
	duration time.Duration
	name     string
}

func (shift *mentorShift) time(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, shift.hour, 0, 0, 0, time.Local)
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

func (ms mentorShifts) getMentorsOnDuty() (mentorsOnDuty []mentorShift) {
	return ms.getShiftsAtTime(time.Now())
}

func (ms mentorShifts) getShiftsAtTime(t time.Time) (shifts []mentorShift) {
	for _, shift := range ms.getShiftsOnWeekday(t.Weekday()) {
		shiftStart := shift.time(t.Date())
		// In the shift or right at the start of it
		if (shiftStart.Before(t) && shiftStart.Add(shift.duration).After(t)) || shiftStart == t {
			shifts = append(shifts, shift)
		}
	}
	return
}

func (ms mentorShifts) getShiftsOnWeekday(weekday time.Weekday) (shifts []mentorShift) {
	for _, shift := range ms {
		if shift.weekday == weekday {
			shifts = append(shifts, shift)
		}
	}
	return
}

func (ms mentorShifts) getShiftsAfterTime(t time.Time) (shifts []mentorShift) {
	for _, shift := range ms.getShiftsOnWeekday(t.Weekday()) {
		shiftStart := shift.time(t.Date())
		if shiftStart.After(t) {
			shifts = append(shifts, shift)
		}
	}
	return
}

func (ms mentorShifts) getNextMentorsOnDutyToday() (shifts []mentorShift) {
	now := time.Now()
	for _, shift := range ms.getShiftsAfterTime(now) {
		shiftStart := shift.time(now.Date())
		shifts = ms.getShiftsAtTime(shiftStart)
		return
	}
	return
}
