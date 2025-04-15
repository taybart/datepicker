package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

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
			"sunday": {
				Short:   "s",
				Help:    "Start week on sunday",
				Default: false,
			},
			"full": {
				Short:   "f",
				Help:    "fullscreen prompt",
				Default: false,
			},
			"prompt": {
				Short:   "p",
				Help:    "title of the calendar prompt",
				Default: "",
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
	quit       bool
	selected   bool
	rangeStart time.Time
	prompt     string
	cal        Calendar
	keys       KeyMap
	help       help.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quit = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Cancel):
			if !m.rangeStart.IsZero() { // stop selection first
				m.rangeStart = time.Time{}
			} else {
				m.quit = true
				return m, tea.Quit
			}
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
		case key.Matches(msg, m.keys.StartRange):
			m.rangeStart = m.cal.date
		case key.Matches(msg, m.keys.Select):
			m.selected = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.selected || m.quit {
		return "" // clear output
	}
	var s strings.Builder

	if m.prompt != "" {
		prompt := m.prompt
		if len(prompt) < 20 {
			padLen := 20 - len(prompt)
			prompt = fmt.Sprintf("%s%s%s",
				strings.Repeat(" ", (padLen+1)/2),
				prompt,
				strings.Repeat(" ", (padLen+1)/2),
			)

		}
		s.WriteString(lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(lipgloss.Color("10")).
			Render(prompt))
		s.WriteString("\n")

	}

	s.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Render(m.cal.MonthHeader()))
	s.WriteString("\n")

	s.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Render(m.cal.WeekHeader()))
	s.WriteString("\n")

	month := m.cal.Map()
	behind := false
	after := false
	selecting := false
	oneMore := false
	for _, week := range month {
		for k, day := range week {
			if day == -1 {
				s.WriteString("   ")
				continue
			}

			dateDay := day + 1 // 0 indexed days
			isWeekend := k >= 5
			focused := dateDay == m.cal.Day()
			currentDate := time.Date(m.cal.Year(), m.cal.Month(), dateDay, 0, 0, 0, 0, time.UTC)

			// there is a range started
			if !m.rangeStart.IsZero() {
				// the date we are checking is the start of selection
				if currentDate.Equal(m.rangeStart) {
					// start selecting at the range select
					selecting = true
					// we actually want to stop selecting since the day we
					// have selected is behind the selection start
					if behind {
						selecting = false
						behind = false
						oneMore = true // we need to show all the way up to the range start day
					}
					// we are focused on the range start, no need to select any more dates
					if focused {
						selecting = false
					}
					// we are at the focused day and our selection starts after this
				} else if focused && currentDate.Before(m.rangeStart) {
					behind = true // we need to stop when we hit the start of range selection
					selecting = true
					// we are at the focused day and our selection starts before this
				} else if focused && currentDate.After(m.rangeStart) {
					after = true
					selecting = false
					// we are in a month in the future and not after the focused day
					// so start selecting up until then
				} else if !after && currentDate.Month()-m.rangeStart.Month() > 0 {
					selecting = true
				}
			}

			style := lipgloss.NewStyle()
			if selecting || oneMore {
				style = style.Background(lipgloss.Color("4")).Foreground(lipgloss.Color("0"))
				oneMore = false
			}
			if m.cal.IsToday(dateDay) {
				style = style.Foreground(lipgloss.Color("9"))
				if !selecting && focused {
					style = style.Background(lipgloss.Color("9")).Foreground(lipgloss.Color("0"))
				}
			} else if isWeekend {
				style = style.Foreground(lipgloss.Color("4"))
				if !selecting && focused {
					style = style.Background(lipgloss.Color("4")).Foreground(lipgloss.Color("0"))
				}
				if selecting {
					style = style.Foreground(lipgloss.Color("15"))
				}
			} else {
				style = style.Foreground(lipgloss.Color("3"))
				if !selecting && focused {
					style = style.Background(lipgloss.Color("3")).Foreground(lipgloss.Color("0"))
				}
			}

			s.WriteString(style.Render(fmt.Sprintf("%2d ", dateDay)))
		}

		s.WriteString("\n")
	}

	s.WriteString("\n")
	if len(month) == 4 { // padding if month fits into 4 weeks exactly
		s.WriteString("\n")
	}
	s.WriteString(m.help.View(m.keys))
	return s.String()
}

func run() error {
	if err := app.Parse(); err != nil {
		if errors.Is(err, args.ErrUsageRequested) {
			return nil
		}
		return err
	}

	cal := NewCalendar()
	cal.SetOutputFormat(app.String("output"))
	cal.SundayStart(app.Bool("sunday"))

	// tmp fix for lipgloss not detecting color output
	os.Setenv("CLICOLOR_FORCE", "true")

	opts := []tea.ProgramOption{tea.WithOutput(os.Stderr)}
	if app.True("full") {
		opts = append(opts, tea.WithAltScreen())
	}

	tm, err := tea.NewProgram(model{
		selected: false,
		prompt:   app.String("prompt"),
		cal:      cal,
		keys:     Keys,
		help:     help.New(),
	}, opts...).Run()
	if err != nil {
		return err
	}

	if m, ok := tm.(model); ok && m.selected {
		if !m.rangeStart.IsZero() {
			if m.rangeStart.After(m.cal.Current()) {
				fmt.Println(m.cal.Format(m.cal.Current()), m.cal.Format(m.rangeStart))
			} else {
				fmt.Println(m.cal.Format(m.rangeStart), m.cal.Format(m.cal.Current()))
			}
			return nil
		}
		fmt.Println(m.cal.Format(m.cal.Current()))
		return nil
	}
	return errors.New("no date picked")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
