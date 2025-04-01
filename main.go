package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/taybart/args"
)

var (
	app = args.App{
		Name:    "cal",
		Version: "0.0.1",
		Author:  "taybart",
		About:   "TUI date picker",
		Args: map[string]*args.Arg{
			// TODO: this
			"startSunday": {
				Short:   "s",
				Help:    "Start week on sunday",
				Default: false,
			},
			"output": {
				Short:   "o",
				Help:    "output format, uses go date formatting (https://pkg.go.dev/time#example-Time.Format)",
				Default: "2006-01-02",
			},
		},
	}
)

type model struct {
	selected bool
	cal      Calendar
	keys     KeyMap
	help     help.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Today):
			m.cal.Today()
		case key.Matches(msg, m.keys.Left):
			m.cal.AddDay(-1)
		case key.Matches(msg, m.keys.Right):
			m.cal.AddDay(1)
		case key.Matches(msg, m.keys.Down):
			m.cal.AddDay(7)
		case key.Matches(msg, m.keys.Up):
			m.cal.AddDay(-7)
		case key.Matches(msg, m.keys.WeekStart):
			m.cal.WeekStart()
		case key.Matches(msg, m.keys.WeekEnd):
			m.cal.WeekEnd()
		case key.Matches(msg, m.keys.MonthStart):
			m.cal.MonthStart()
		case key.Matches(msg, m.keys.MonthEnd):
			m.cal.MonthEnd()
		case key.Matches(msg, m.keys.MonthPrev):
			m.cal.AddMonth(-1)
		case key.Matches(msg, m.keys.MonthNext):
			m.cal.AddMonth(1)
		case key.Matches(msg, m.keys.YearPrev):
			m.cal.AddYear(-1)
		case key.Matches(msg, m.keys.YearNext):
			m.cal.AddYear(1)
		case key.Matches(msg, m.keys.Select):
			m.selected = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.selected {
		return "" // clear output
	}

	s := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Render(fmt.Sprintf("   %s %d\nMo Tu We Th Fr Sa Su", m.cal.Month(), m.cal.Year())) + "\n"

	monthMap := m.cal.Map()
	for _, week := range monthMap {
		for k, day := range week {
			if day == 0 {
				s += "   "
				continue
			}

			isWeekend := k >= 5
			focused := day == m.cal.Day()
			style := lipgloss.NewStyle()

			if m.cal.IsToday(day) {
				if focused {
					style = style.Background(lipgloss.Color("9")).Foreground(lipgloss.Color("0"))
				} else {
					style = style.Foreground(lipgloss.Color("9"))
				}
			} else if isWeekend {
				if focused {
					style = style.Background(lipgloss.Color("4")).Foreground(lipgloss.Color("0"))
				} else {
					style = style.Foreground(lipgloss.Color("4"))
				}
			} else {
				if focused {
					style = style.Background(lipgloss.Color("3")).Foreground(lipgloss.Color("0"))
				} else {
					style = style.Foreground(lipgloss.Color("3"))
				}
			}
			s += style.Render(fmt.Sprintf("%2d ", day))
		}

		s += "\n"
	}

	if len(monthMap) == 4 {
		s += "\n\n"
	} else if len(monthMap) == 5 {
		s += "\n"
	}
	s += m.help.View(m.keys)
	return s
}

func run() error {
	if err := app.Parse(); err != nil {
		return err
	}

	cal := NewCalendar()
	cal.SetOutputFormat(app.String("output"))

	tm, err := tea.NewProgram(model{
		selected: false,
		cal:      cal,
		keys:     Keys,
		help:     help.New(),
	}, tea.WithOutput(os.Stderr)).Run()
	if err != nil {
		return err
	}

	if m, ok := tm.(model); ok && m.selected {
		fmt.Println(m.cal.Current())
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
