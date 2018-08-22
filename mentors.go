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

type mentorShifts [16]mentorShift

const mentorDefaultShiftDuration = time.Duration(time.Hour * 2)

// Mentor names array. Each row is a day of the week (sun, mon, ..., sat). Each element in a
// row is a mentor timeslot starting at 12PM, where each slot is 2 hours long.
var shifts = mentorShifts{
	{16, time.Sunday, mentorDefaultShiftDuration, "Brayden A"},
	{18, time.Sunday, mentorDefaultShiftDuration, "Jesse L"},
	{14, time.Monday, mentorDefaultShiftDuration, "David L"},
	{16, time.Monday, mentorDefaultShiftDuration, "Emily Mc."},
	{18, time.Monday, mentorDefaultShiftDuration, "Josh P"},
	{14, time.Tuesday, mentorDefaultShiftDuration, "Sabina S"},
	{16, time.Tuesday, mentorDefaultShiftDuration, "Braden B"},
	{18, time.Tuesday, mentorDefaultShiftDuration, "Christina H"},
	{20, time.Tuesday, mentorDefaultShiftDuration, "Nick B"},
	{14, time.Wednesday, mentorDefaultShiftDuration, "Liam K"},
	{16, time.Wednesday, mentorDefaultShiftDuration, "Swapnil P"},
	{18, time.Wednesday, mentorDefaultShiftDuration, "Sameer P"},
	{20, time.Wednesday, mentorDefaultShiftDuration, "Alex B"},
	{14, time.Thursday, mentorDefaultShiftDuration, "Lin L"},
	{16, time.Thursday, mentorDefaultShiftDuration, "Olivia C"},
	{18, time.Thursday, mentorDefaultShiftDuration, "Emily Mar."},
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
