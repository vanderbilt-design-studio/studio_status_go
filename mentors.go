package main

import "time"

type mentorShift struct {
	hour     int
	min      int
	weekday  time.Weekday
	duration time.Duration
	name     string
}

func (shift *mentorShift) time(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, shift.hour, shift.min, 0, 0, time.Local)
}

type mentorShifts [29]mentorShift

const mentorDefaultShiftDuration = time.Duration(time.Hour * 2)

// Mentor names array. Each row is a day of the week (sun, mon, ..., sat). Each element in a
// row is a mentor timeslot starting at 12PM, where each slot is 2 hours long.
var shifts = mentorShifts{
	{16, 0, time.Sunday, mentorDefaultShiftDuration, "Brayden A"},
	{17, 0, time.Sunday, mentorDefaultShiftDuration, "Sofia R"},
	{18, 0, time.Sunday, mentorDefaultShiftDuration, "Edward D"},
	{20, 0, time.Sunday, mentorDefaultShiftDuration, "Xue Ye L"},
	{14, 0, time.Monday, mentorDefaultShiftDuration, "Christina H"},
	{16, 0, time.Monday, mentorDefaultShiftDuration, "Sabina S"},
	{18, 0, time.Monday, mentorDefaultShiftDuration, "Lin L"},
	{20, 0, time.Monday, mentorDefaultShiftDuration, "Tristan I"},
	{12, 0, time.Tuesday, mentorDefaultShiftDuration, "Joey H"},
	{16, 0, time.Tuesday, mentorDefaultShiftDuration, "Sophia Z"},
	{17, 0, time.Tuesday, mentorDefaultShiftDuration, "Sam S"},
	{18, 0, time.Tuesday, mentorDefaultShiftDuration, "Emily Mc"},
	{20, 0, time.Tuesday, mentorDefaultShiftDuration, "Diandry R"},
	{12, 0, time.Wednesday, mentorDefaultShiftDuration, "Sameer P"},
	{14, 0, time.Wednesday, mentorDefaultShiftDuration, "Swapnil P"},
	{16, 0, time.Wednesday, mentorDefaultShiftDuration, "Amaury P"},
	{17, 0, time.Wednesday, mentorDefaultShiftDuration, "Zach S"},
	{18, 0, time.Wednesday, mentorDefaultShiftDuration, "David L"},
	{19, 0, time.Wednesday, mentorDefaultShiftDuration, "Paolo D"},
	{20, 0, time.Wednesday, mentorDefaultShiftDuration, "Josh P"},
	{14, 30, time.Thursday, mentorDefaultShiftDuration, "Jason Y"},
	{16, 0, time.Thursday, mentorDefaultShiftDuration, "Olivia C"},
	{18, 0, time.Thursday, mentorDefaultShiftDuration, "Emily Mar."},
	{19, 0, time.Thursday, mentorDefaultShiftDuration, "Patia F"},
	{20, 0, time.Thursday, mentorDefaultShiftDuration, "Amy C"},
	{12, 0, time.Friday, mentorDefaultShiftDuration, "Alex S"},
	{13, 0, time.Friday, mentorDefaultShiftDuration, "Will R"},
	{14, 0, time.Friday, mentorDefaultShiftDuration, "Jack M"},
	{16, 0, time.Friday, mentorDefaultShiftDuration, "Nick B"},
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
