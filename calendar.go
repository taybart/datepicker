package main

import (
	"fmt"
	"time"
)

type Calendar struct {
	date   time.Time
	format string
}
type Month []Week
type Week [7]int

func (week Week) firstDay() int {
	for _, day := range week {
		if day != 0 {
			return day
		}
	}
	return 0
}
func (week Week) lastDay() int {
	for i := 6; i >= 0; i-- {
		if week[i] != 0 {
			return week[i]
		}
	}
	return 0
}

func NewCalendar() Calendar {
	return Calendar{
		date:   time.Now(),
		format: "2006-01-02",
	}
}

func (c *Calendar) SetOutputFormat(format string) {
	// TODO: check if format is valid, i don't think this is possible
	c.format = format
}

/*
 * Xetters
 */
func (c *Calendar) Today() time.Time {
	c.date = time.Now()
	return c.date
}
func (c *Calendar) IsToday(day int) bool {
	now := time.Now()
	return day == time.Now().Day() && c.Month() == now.Month() && c.Year() == now.Year()
}
func (c Calendar) Current() string {
	return c.date.Format(fmt.Sprintf("%s\n", c.format))
}
func (c Calendar) Day() int {
	return c.date.Day()
}
func (c Calendar) Month() time.Month {
	return c.date.Month()
}
func (c Calendar) Year() int {
	return c.date.Year()
}
func (c Calendar) lastDay() int {
	return time.Date(c.Year(), c.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func (c Calendar) firstWeekdayOfMonth() int {
	weekday := time.Date(c.Year(), c.Month(), 1, 0, 0, 0, 0, time.UTC).Weekday()
	// convert to monday as first of week
	return (int(weekday) + 5) % 7
}

/*
 * Operations
 */
func (c *Calendar) SetDate(newDate time.Time) {
	c.date = newDate
}
func (c *Calendar) AddDay(ammount int) {
	c.date = c.date.AddDate(0, 0, ammount)
}
func (c *Calendar) AddMonth(ammount int) {
	c.date = c.date.AddDate(0, 0, ammount)
}
func (c *Calendar) AddYear(ammount int) {
	c.date = c.date.AddDate(0, 0, ammount)
}
func (c *Calendar) WeekStart() {
	d := c.Day() - c.Map()[c.week()].firstDay()
	c.AddDay(-d)
}
func (c *Calendar) WeekEnd() {
	d := c.Map()[c.week()].lastDay() - c.Day()
	c.AddDay(d)
}
func (c *Calendar) MonthStart() {
	c.date = time.Date(c.date.Year(), c.date.Month(), 1, 0, 0, 0, 0, time.UTC)
}
func (c *Calendar) MonthEnd() {
	c.date = time.Date(c.Year(), c.Month(), c.lastDay(), 0, 0, 0, 0, time.UTC)
}

// TODO what is this?
func (c Calendar) Map() Month {
	monthMap := make(Month, 0)
	week := Week{}

	// fill weeks of the month
	startDay := c.firstWeekdayOfMonth()
	for day := range c.lastDay() + 1 {
		week[startDay%7] = day
		startDay += 1
		if startDay%7 == 0 {
			monthMap = append(monthMap, week)
			week = Week{}
		}
	}
	// is this a 5 week month?
	if startDay%7 > 0 {
		monthMap = append(monthMap, week)
	}
	return monthMap
}

// TODO fix
func (c Calendar) week() int {
	firstWeekday := c.firstWeekdayOfMonth()
	if firstWeekday == 0 {
		firstWeekday = 7
	}
	return (c.Day() + int(firstWeekday-2)) / 7
}
