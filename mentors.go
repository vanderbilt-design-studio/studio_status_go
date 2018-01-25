package main

import "time"

type mentor_shift struct {
	start    time.Time
	duration time.Duration
	name     string
}

type mentor_shifts []mentor_shift

func MustParse(layout, value string) time.Time {
	if t, err := time.Parse(layout, value); err != nil {
		panic(err)
	} else {
		return t
	}
}

const mentorTimeLayout = "Mon 3:00PM"
const mentorDefaultShiftDuration = time.Duration(time.Hour * 2)

// Mentor names array. Each row is a day of the week (sun, mon, ..., sat). Each element in a
// row is a mentor timeslot starting at 12PM, where each slot is 2 hours long.
var shifts = mentor_shifts{
	{MustParse(mentorTimeLayout, "Sun 2:00PM"), mentorDefaultShiftDuration, "Iliya L"},
	{MustParse(mentorTimeLayout, "Sun 4:00PM"), mentorDefaultShiftDuration, "Nick B"},
	{MustParse(mentorTimeLayout, "Sun 6:00PM"), mentorDefaultShiftDuration, "Foard N"},
	{MustParse(mentorTimeLayout, "Mon 2:00PM"), mentorDefaultShiftDuration, "Lin L"},
	{MustParse(mentorTimeLayout, "Mon 4:00PM"), mentorDefaultShiftDuration, "Amaury P"},
	{MustParse(mentorTimeLayout, "Mon 6:00PM"), mentorDefaultShiftDuration, "Jeremy D"},
	{MustParse(mentorTimeLayout, "Mon 8:00PM"), mentorDefaultShiftDuration, "Kurt L"},
	{MustParse(mentorTimeLayout, "Tue 2:00PM"), mentorDefaultShiftDuration, "Sophia Z"},
	{MustParse(mentorTimeLayout, "Tue 4:00PM"), mentorDefaultShiftDuration, "Emily Mk"},
	{MustParse(mentorTimeLayout, "Tue 6:00PM"), mentorDefaultShiftDuration, "Jonah H"},
	{MustParse(mentorTimeLayout, "Wed 2:00PM"), mentorDefaultShiftDuration, "Eric N"},
	{MustParse(mentorTimeLayout, "Wed 4:00PM"), mentorDefaultShiftDuration, "Lauren B"},
	{MustParse(mentorTimeLayout, "Wed 6:00PM"), mentorDefaultShiftDuration, "Sameer P"},
	{MustParse(mentorTimeLayout, "Thu 8:00PM"), mentorDefaultShiftDuration, "Christina H"},
	{MustParse(mentorTimeLayout, "Thu 2:00PM"), mentorDefaultShiftDuration, "Alex B"},
	{MustParse(mentorTimeLayout, "Thu 4:00PM"), mentorDefaultShiftDuration, "Emily Mc"},
	{MustParse(mentorTimeLayout, "Thu 6:00PM"), mentorDefaultShiftDuration, "Braden B"},
	{MustParse(mentorTimeLayout, "Thu 8:00PM"), mentorDefaultShiftDuration, "Jill B"},
	{MustParse(mentorTimeLayout, "Fri 2:00PM"), mentorDefaultShiftDuration, "Dominic G"},
	{MustParse(mentorTimeLayout, "Fri 4:00PM"), mentorDefaultShiftDuration, "Josh P"},

	{MustParse(mentorTimeLayout, "Tue 3:00PM"), mentorDefaultShiftDuration, "David L"},
	{MustParse(mentorTimeLayout, "Tue 5:00PM"), mentorDefaultShiftDuration, "Yunyu L"},
	{MustParse(mentorTimeLayout, "Wed 5:00PM"), mentorDefaultShiftDuration, "Swapnil P"},
	{MustParse(mentorTimeLayout, "Thu 3:00PM"), mentorDefaultShiftDuration, "Joey H"},
	{MustParse(mentorTimeLayout, "Thu 5:00PM"), mentorDefaultShiftDuration, "Jesse L"},
}

func (this mentor_shifts) getMentorsOnDuty() (mentorOnDuty []mentor_shift) {
	now := time.Now()
	y, m, d := now.AddDate(0, 0, -int(now.Weekday())).Date()
	mentorsOnDuty := make([]mentor_shift, 0, 2)
	for _, shift := range this {
		shiftStart := shift.start.AddDate(y, int(m), d)
		if shiftStart.Before(now) && shiftStart.Add(shift.duration).After(now) {
			mentorsOnDuty = append(mentorsOnDuty, shift)
		}
	}
	return
}
