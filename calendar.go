package main

import (
	"fmt"
	"time"
)

type Calendar struct {
	date        time.Time
	format      string
	startSunday bool
}
type Month []Week
type Week [7]int

func NewWeek() Week {
	return Week{-1, -1, -1, -1, -1, -1, -1}
}

func (week Week) firstDay() int {
	for _, day := range week {
		if day != -1 {
			return day + 1
		}
	}
	return 0
}
func (week Week) lastDay() int {
	for i := 6; i >= 0; i-- {
		if week[i] != -1 {
			return week[i] + 1
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
func (c *Calendar) SundayStart(s bool) {
	c.startSunday = s
}

/*
 * Render
 */
func (c Calendar) MonthHeader() string {
	return fmt.Sprintf("   %s %d", c.Month(), c.Year())
}
func (c Calendar) WeekHeader() string {
	if c.startSunday {
		return "Su Mo Tu We Th Fr Sa"
	}
	return "Mo Tu We Th Fr Sa Su"
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
	// return c.date.Format(fmt.Sprintf("%s\n", c.format))
	return c.date.Format(c.format)
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
func (c Calendar) lastDayOfMonth() int {
	return time.Date(c.Year(), c.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func (c Calendar) firstWeekdayOfMonth() int {
	weekday := time.Date(c.Year(), c.Month(), 1, 0, 0, 0, 0, time.UTC).Weekday()
	if c.startSunday {
		return int(weekday)
	}
	// convert to monday as first of week
	return int(weekday+6) % 7
}

/*
 * Operations
 */
func (c *Calendar) SetDate(newDate time.Time) {
	c.date = newDate
}
func (c *Calendar) AddDay(amount int) {
	c.date = c.date.AddDate(0, 0, amount)
}
func (c *Calendar) AddMonth(amount int) {
	c.date = c.date.AddDate(0, amount, 0)
}
func (c *Calendar) AddYear(amount int) {
	c.date = c.date.AddDate(amount, 0, 0)
}
func (c *Calendar) WeekStart() {
	d := c.Map()[c.weekIndex()].firstDay()
	c.date = time.Date(c.date.Year(), c.date.Month(), d, 0, 0, 0, 0, time.UTC)
}
func (c *Calendar) WeekEnd() {
	d := c.Map()[c.weekIndex()].lastDay()
	c.date = time.Date(c.date.Year(), c.date.Month(), d, 0, 0, 0, 0, time.UTC)
}
func (c *Calendar) MonthStart() {
	c.date = time.Date(c.date.Year(), c.date.Month(), 1, 0, 0, 0, 0, time.UTC)
}
func (c *Calendar) MonthEnd() {
	c.date = time.Date(c.Year(), c.Month(), c.lastDayOfMonth(), 0, 0, 0, 0, time.UTC)
}

func (c Calendar) Map() Month {
	monthMap := make(Month, 0)
	week := NewWeek()

	// fill weeks of the month
	startDay := c.firstWeekdayOfMonth()
	for day := range c.lastDayOfMonth() {
		week[startDay%7] = day
		startDay += 1
		if startDay%7 == 0 {
			monthMap = append(monthMap, week)
			week = NewWeek()
		}
	}
	// is this a 5 week month?
	if startDay%7 > 0 {
		monthMap = append(monthMap, week)
	}
	return monthMap
}

// TODO fix
func (c Calendar) weekIndex() int {
	return (c.firstWeekdayOfMonth() + c.Day() - 1) / 7
}
