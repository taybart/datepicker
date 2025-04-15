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
	lg "github.com/charmbracelet/lipgloss"
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
	width      int
	height     int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
		s.WriteString(lg.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(lg.Color("10")).
			Render(prompt))
		s.WriteString("\n")

	}

	s.WriteString(lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("5")).
		Render(m.cal.MonthHeader()))
	s.WriteString("\n")

	s.WriteString(lg.NewStyle().
		Bold(true).
		Foreground(lg.Color("5")).
		Render(m.cal.WeekHeader()))
	s.WriteString("\n")

	month := m.cal.Map()
	for _, week := range month {
		for k, day := range week {
			if day == -1 {
				s.WriteString("   ")
				continue
			}

			// Basic date information
			dateDay := day + 1 // 0 indexed days
			isWeekend := k >= 5
			currentDate := time.Date(m.cal.Year(), m.cal.Month(), dateDay, 0, 0, 0, 0, time.UTC)

			// Focus information
			isFocused := dateDay == m.cal.Day()
			focusedDate := time.Date(m.cal.Year(), m.cal.Month(), m.cal.Day(), 0, 0, 0, 0, time.UTC)

			// Range selection logic
			isSelected := false
			isRangeEndpoint := false

			// Only process range selection if a range start exists
			if !m.rangeStart.IsZero() {
				isRangeStartDay := currentDate.Equal(m.rangeStart)

				// Case 1: Selecting forward (focus to range start)
				if focusedDate.Before(m.rangeStart) {
					// Select days between focus and range start
					isInForwardRange := currentDate.After(focusedDate) && currentDate.Before(m.rangeStart)
					isSelected = isFocused || isInForwardRange
					isRangeEndpoint = isRangeStartDay
					// Case 2: Selecting backward (range start to focus)
				} else if focusedDate.After(m.rangeStart) {
					// Select days between range start and focus
					isInBackwardRange := currentDate.After(m.rangeStart) && currentDate.Before(focusedDate)
					isSelected = isRangeStartDay || isInBackwardRange
					// Case 3: Focus is on range start
				} else if focusedDate.Equal(m.rangeStart) {
					// Only select the range start if it's not focused
					isSelected = isRangeStartDay && !isFocused
				}

				// Handle multi-month selection
				if !focusedDate.Equal(m.rangeStart) &&
					currentDate.Month() != focusedDate.Month() &&
					currentDate.Month() != m.rangeStart.Month() {

					// For multi-month forward selection
					if focusedDate.Before(m.rangeStart) &&
						currentDate.After(focusedDate) &&
						currentDate.Before(m.rangeStart) {
						isSelected = true
						// For multi-month backward selection
					} else if focusedDate.After(m.rangeStart) &&
						currentDate.After(m.rangeStart) &&
						currentDate.Before(focusedDate) {
						isSelected = true
					}
				}
			}
			// Styling logic
			style := lg.NewStyle()

			// Apply selection styling
			if isSelected || isRangeEndpoint {
				style = style.Background(lg.Color("4")).Foreground(lg.Color("0"))
			}

			// Apply other styling based on day type
			if m.cal.IsToday(dateDay) {
				style = style.Foreground(lg.Color("9"))
				if !isSelected && isFocused {
					style = style.Background(lg.Color("9")).Foreground(lg.Color("0"))
				}
			} else if isWeekend {
				style = style.Foreground(lg.Color("4"))
				if !isSelected && isFocused {
					style = style.Background(lg.Color("4")).Foreground(lg.Color("0"))
				}
				if isSelected {
					style = style.Foreground(lg.Color("15"))
				}
			} else {
				style = style.Foreground(lg.Color("3"))
				if !isSelected && isFocused {
					style = style.Background(lg.Color("3")).Foreground(lg.Color("0"))
				}
			}

			s.WriteString(style.Render(fmt.Sprintf("%2d ", dateDay)))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	// pad one more line if month fits into 4 weeks exactly
	if len(month) == 4 {
		s.WriteString("\n")
	}
	// show help keys
	s.WriteString(m.help.View(m.keys))

	if app.True("full") {
		calendar := strings.Split(s.String(), "\n")

		maxWidth := 20
		// Create a centered style
		centered := lg.NewStyle().
			Width(maxWidth).
			PaddingLeft((m.width - maxWidth) / 2)

		// Apply the centered style to each line
		var s strings.Builder
		for _, line := range calendar {
			s.WriteString(centered.Render(line))
			s.WriteString("\n")
		}

		// Position vertically using padding
		verticalPadding := (m.height - len(calendar)) / 2
		if verticalPadding > 0 {
			return strings.Repeat("\n", verticalPadding) + s.String()
		}
		return s.String()
	}
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

	// tmp (but really permanent) fix for lg not detecting color output
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
