package main

import (
	"flag"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
)

var bigDigits = map[rune][]string{
	'0': {
		"██████",
		"█    █",
		"█    █",
		"█    █",
		"██████",
	},
	'1': {
		"  ██  ",
		"  ██  ",
		"  ██  ",
		"  ██  ",
		"  ██  ",
	},
	'2': {
		"██████",
		"     █",
		"██████",
		"█     ",
		"██████",
	},
	'3': {
		"██████",
		"     █",
		"██████",
		"     █",
		"██████",
	},
	'4': {
		"█    █",
		"█    █",
		"██████",
		"     █",
		"     █",
	},
	'5': {
		"██████",
		"█     ",
		"██████",
		"     █",
		"██████",
	},
	'6': {
		"██████",
		"█     ",
		"██████",
		"█    █",
		"██████",
	},
	'7': {
		"██████",
		"     █",
		"     █",
		"     █",
		"     █",
	},
	'8': {
		"██████",
		"█    █",
		"██████",
		"█    █",
		"██████",
	},
	'9': {
		"██████",
		"█    █",
		"██████",
		"     █",
		"██████",
	},
	':': {
		"      ",
		"  ██  ",
		"      ",
		"  ██  ",
		"      ",
	},
}

var (
	colorBlue   = lipgloss.Color("33")
	colorYellow = lipgloss.Color("220")
	colorSubtle = lipgloss.Color("#999999")

	styleContainer = lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center)
	styleInput     = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(colorSubtle).Padding(1, 3).Width(40)
	styleHelp      = lipgloss.NewStyle().Foreground(colorSubtle).MarginTop(3)
)

type sessionState int

const (
	stateSetup sessionState = iota
	stateRunning
	statePopup
)

type timerType int

const (
	typeWork timerType = iota
	typeBreak
)

type model struct {
	width  int
	height int

	state     sessionState
	timerType timerType
	paused    bool

	inputs     []textinput.Model
	focusIndex int

	workDuration  time.Duration
	breakDuration time.Duration
	timeLeft      time.Duration

	sessionsTotal  int
	currentSession int

	timerID   int
	autobreak bool

	popupMessage string
	popupCallback func(bool)
}


func initialModel(workArg, breakArg, sessArg string) model {
	m := model{
		inputs:    make([]textinput.Model, 3),
		timerID:   0,
		autobreak: true,
	}

	t0 := textinput.New()
	t0.Placeholder = "Work (e.g. 25, 30s)"
	t0.Focus()
	t0.Width = 30
	t1 := textinput.New()
	t1.Placeholder = "Break (e.g. 5m)"
	t1.Width = 30
	t2 := textinput.New()
	t2.Placeholder = "Sessions (e.g. 4)"
	t2.Width = 30

	m.inputs[0] = t0
	m.inputs[1] = t1
	m.inputs[2] = t2

	if workArg != "" {
		m.state = stateRunning
		m.timerType = typeWork
		m.paused = false
		m.currentSession = 1
		m.workDuration = parseDurationInput(workArg, 25)
		m.breakDuration = parseDurationInput(breakArg, 5)
		s, _ := strconv.Atoi(sessArg)
		if s == 0 {
			s = 4
		}
		m.sessionsTotal = s
		m.timeLeft = m.workDuration

		m.timerID++
	} else {
		m.state = stateSetup
		m.timerType = typeWork
	}

	return m
}

func (m model) Init() tea.Cmd {
		if m.state == stateRunning {
		return tea.Batch(textinput.Blink, doTick(m.timerID))
	}
	return textinput.Blink
}


type tickMsg struct {
	id int
}

func doTick(id int) tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{id: id}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Model updated with new dimensions, View() will be called automatically
		return m, nil

	case tickMsg:
		if msg.id != m.timerID {
			return m, nil
		}

		if m.state == stateRunning && !m.paused && m.timeLeft > 0 {
			m.timeLeft -= time.Second
			if m.timeLeft <= 0 {
				return m.handleTimerFinish()
			}
			return m, doTick(m.timerID) // <--- CHANGED: Pass ID
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		if m.state == stateSetup {
			switch msg.String() {
			case "a", "A":
				m.autobreak = !m.autobreak
				return m, nil
			case "tab", "shift+tab", "enter", "up", "down":
				s := msg.String()
				if s == "enter" && m.focusIndex == len(m.inputs)-1 {
					return m.startTimer()
				}
				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}
				if m.focusIndex > len(m.inputs)-1 {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = len(m.inputs) - 1
				}

				cmds := make([]tea.Cmd, len(m.inputs))
				for i := 0; i <= len(m.inputs)-1; i++ {
					if i == m.focusIndex {
						cmds[i] = m.inputs[i].Focus()
						m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(colorBlue)
					} else {
						m.inputs[i].Blur()
						m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(colorSubtle)
					}
				}
				return m, tea.Batch(cmds...)
			}
		}

		if m.state == stateRunning {
			switch msg.String() {
			case " ":
				m.paused = !m.paused
				if !m.paused {
					return m, doTick(m.timerID)
				}
			case "s":
				return m.handleTimerFinish()
			case "up":
				m.timeLeft += time.Minute
			case "down":
				if m.timeLeft > time.Minute {
					m.timeLeft -= time.Minute
				}
			case "esc":
				return m.resetToSetup()
			}
		}

		if m.state == statePopup {
			switch msg.String() {
			case "y", "Y":
				if strings.Contains(m.popupMessage, "Work session") {
					// Start break
					msg := fmt.Sprintf("Work session %d/%d finished! Time for a break.", m.currentSession, m.sessionsTotal)
					_ = beeep.Notify("Pomodoro", msg, "")
					m.timerType = typeBreak
					m.timeLeft = m.breakDuration
					m.paused = false
					m.state = stateRunning
					return m, doTick(m.timerID)
				} else {
					// Start work
					msg := fmt.Sprintf("Break finished! Starting work session %d/%d.", m.currentSession+1, m.sessionsTotal)
					_ = beeep.Notify("Pomodoro", msg, "")
					m.timerType = typeWork
					m.timeLeft = m.workDuration
					m.currentSession++
					m.paused = false
					m.state = stateRunning

					if m.currentSession > m.sessionsTotal {
						_ = beeep.Notify("Pomodoro", fmt.Sprintf("All %d sessions completed! Great work!", m.sessionsTotal), "")
						return m, tea.Quit
					}
					return m, doTick(m.timerID)
				}
			case "n", "N":
				// Stay in current state, don't start next timer
				m.state = stateRunning
				return m, nil
			case "q":
				return m, tea.Quit
			}
		}
	}

	if m.state == stateSetup {
		cmd := m.updateInputs(msg)
		return m, cmd
	}

	return m, nil
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}


func parseDurationInput(s string, defaultMin int) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Duration(defaultMin) * time.Minute
	}
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	if val, err := strconv.Atoi(s); err == nil {
		return time.Duration(val) * time.Minute
	}
	return time.Duration(defaultMin) * time.Minute
}

func playWindowsSound() {
	if runtime.GOOS == "windows" {
		go func() {
			_ = exec.Command("powershell", "-c", "(New-Object Media.SoundPlayer 'C:\\Windows\\Media\\Windows Notify System Generic.wav').PlaySync()").Run()
		}()
	} else {
		go beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	}
}

func (m model) startTimer() (model, tea.Cmd) {
	m.workDuration = parseDurationInput(m.inputs[0].Value(), 25)
	m.breakDuration = parseDurationInput(m.inputs[1].Value(), 5)
	s, _ := strconv.Atoi(m.inputs[2].Value())
	if s == 0 {
		s = 4
	}
	m.sessionsTotal = s
	m.currentSession = 1

	m.state = stateRunning
	m.timerType = typeWork
	m.timeLeft = m.workDuration
	m.paused = false

	m.timerID++

	return m, doTick(m.timerID)
}

func (m model) resetToSetup() (model, tea.Cmd) {
	m.state = stateSetup
	m.timerType = typeWork
	m.paused = false
	m.currentSession = 0
	m.timeLeft = 0
	m.timerID++
	return m, nil
}

func (m model) handleTimerFinish() (model, tea.Cmd) {
	playWindowsSound()

	if m.autobreak {
		return m.handleTimerFinishAuto()
	} else {
		return m.handleTimerFinishManual()
	}
}

func (m model) handleTimerFinishAuto() (model, tea.Cmd) {
	m.timerID++

	msg := ""
	if m.timerType == typeWork {
		msg = fmt.Sprintf("Work session %d/%d finished! Time for a break.", m.currentSession, m.sessionsTotal)
		_ = beeep.Notify("Pomodoro", msg, "")
		m.timerType = typeBreak
		m.timeLeft = m.breakDuration
	} else {
		msg = fmt.Sprintf("Break finished! Starting work session %d/%d.", m.currentSession+1, m.sessionsTotal)
		_ = beeep.Notify("Pomodoro", msg, "")
		m.timerType = typeWork
		m.timeLeft = m.workDuration
		m.currentSession++
	}

	if m.currentSession > m.sessionsTotal {
		_ = beeep.Notify("Pomodoro", fmt.Sprintf("All %d sessions completed! Great work!", m.sessionsTotal), "")
		return m, tea.Quit
	}

	m.paused = false
	return m, doTick(m.timerID)
}

func (m model) handleTimerFinishManual() (model, tea.Cmd) {
	m.state = statePopup
	m.timerID++ // Increment ID to stop current timer

	if m.timerType == typeWork {
		m.popupMessage = fmt.Sprintf("Work session %d/%d finished!\nStart the break?", m.currentSession, m.sessionsTotal)
	} else {
		m.popupMessage = fmt.Sprintf("Break finished!\nStart work session %d/%d?", m.currentSession+1, m.sessionsTotal)
	}

	return m, nil
}


func renderBigTime(d time.Duration, color lipgloss.Color, width int) string {
	totalMinutes := int(d.Minutes())

	var timeStr string
	var requiredWidth int

	if totalMinutes >= 60 {
		totalSeconds := int(d.Seconds())
		hours := totalSeconds / 3600
		minutes := (totalSeconds % 3600) / 60
		seconds := totalSeconds % 60
		timeStr = fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
		requiredWidth = 56
	} else {
		seconds := int(d.Seconds()) % 60
		timeStr = fmt.Sprintf("%02d:%02d", totalMinutes, seconds)
		requiredWidth = 35
	}

	// For very small screens (30-20% of typical terminal width), always use simple text
	if width < 40 {
		totalSeconds := int(d.Seconds())
		hours := totalSeconds / 3600
		minutes := (totalSeconds % 3600) / 60
		seconds := totalSeconds % 60
		var smallTimeStr string
		if hours > 0 {
			smallTimeStr = fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
		} else {
			smallTimeStr = fmt.Sprintf("%02d:%02d", minutes, seconds)
		}
		return lipgloss.NewStyle().
			Foreground(color).
			Bold(true).
			Render(smallTimeStr)
	}

	if width < requiredWidth {
		var smallTimeStr string
		if totalMinutes >= 60 {
			totalSeconds := int(d.Seconds())
			hours := totalSeconds / 3600
			minutes := (totalSeconds % 3600) / 60
			seconds := totalSeconds % 60
			smallTimeStr = fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
		} else {
			seconds := int(d.Seconds()) % 60
			smallTimeStr = fmt.Sprintf("%02d:%02d", totalMinutes, seconds)
		}
		return lipgloss.NewStyle().
			Foreground(color).
			Bold(true).
			Render(smallTimeStr)
	}

	height := 5
	lines := make([]string, height)
	for _, char := range timeStr {
		block, ok := bigDigits[char]
		if !ok {
			continue
		}
		for i := 0; i < height; i++ {
			lines[i] += block[i] + " "
		}
	}
	fullBlock := strings.Join(lines, "\n")
	return lipgloss.NewStyle().Foreground(color).Render(fullBlock)
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	var s string
	if m.state == stateSetup {
		s = m.viewSetup()
	} else if m.state == statePopup {
		s = m.viewPopup()
	} else {
		s = m.viewTimer()
	}
	return styleContainer.Width(m.width).Height(m.height).Render(s)
}

func (m model) viewSetup() string {
	var b strings.Builder

	title := "POMODORO SETUP"
	if m.width < 30 {
		title = "SETUP"
	}
	if m.width < 20 {
		title = "CFG"
	}

	var titleStyle lipgloss.Style
	if m.width < 20 {
		titleStyle = lipgloss.NewStyle().Foreground(colorBlue)
	} else {
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	}
	b.WriteString(titleStyle.Render(title) + "\n\n")

	labels := []string{"Work Duration:", "Break Duration:", "Sessions:"}
	if m.width < 40 {
		labels = []string{"Work:", "Break:", "Sessions:"}
	}
	if m.width < 25 {
		labels = []string{"W:", "B:", "S:"}
	}

	for i := 0; i < len(m.inputs); i++ {
		labelStyle := lipgloss.NewStyle().Foreground(colorSubtle)
		if m.width < 20 {
			labelStyle = lipgloss.NewStyle().Foreground(colorSubtle).Bold(false)
		}
		b.WriteString(labelStyle.Render(labels[i]) + "\n")
		b.WriteString(styleInput.Render(m.inputs[i].View()) + "\n\n")
	}

	autobreakStatus := "OFF"
	if m.autobreak {
		autobreakStatus = "ON"
	}
	autobreakLabel := fmt.Sprintf("Autobreak: %s", autobreakStatus)
	if m.width < 30 {
		autobreakLabel = fmt.Sprintf("Auto: %s", autobreakStatus)
	}
	if m.width < 20 {
		autobreakLabel = fmt.Sprintf("A:%s", autobreakStatus)
	}

	labelStyle := lipgloss.NewStyle().Foreground(colorSubtle)
	if m.width < 20 {
		labelStyle = lipgloss.NewStyle().Foreground(colorSubtle).Bold(false)
	}
	b.WriteString(labelStyle.Render(autobreakLabel) + "\n\n")

	if m.width >= 80 {
		helpText := "\n[TAB] Switch  •  [ENTER] Start  •  [a] Autobreak  •  [q] Quit"
		if m.width < 50 {
			helpText = "\n[TAB] Switch  •  [ENTER] Start\n[a] Autobreak  •  [q] Quit"
		}
		if m.width < 30 {
			helpText = "\n[TAB] Sw  •  [ENTER] Go\n[a] Toggle  •  [q] Quit"
		}
		b.WriteString(styleHelp.Render(helpText))
	}
	return b.String()
}

func (m model) viewTimer() string {
	activeColor := colorBlue
	modeStr := fmt.Sprintf("WORK SESSION %d/%d", m.currentSession, m.sessionsTotal)
	if m.timerType == typeBreak {
		activeColor = colorYellow
		modeStr = "BREAK TIME"
	}

	// Make title responsive for narrow terminals
	if m.width < 40 {
		if m.timerType == typeWork {
			modeStr = fmt.Sprintf("WORK %d/%d", m.currentSession, m.sessionsTotal)
		} else {
			modeStr = "BREAK"
		}
	}
	if m.width < 25 {
		if m.timerType == typeWork {
			modeStr = fmt.Sprintf("W%d/%d", m.currentSession, m.sessionsTotal)
		} else {
			modeStr = "BRK"
		}
	}

	// Make title smaller/styled for very narrow terminals
	var titleStyle lipgloss.Style
	if m.width < 25 {
		titleStyle = lipgloss.NewStyle().Foreground(activeColor)
	} else {
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(activeColor)
	}
	title := titleStyle.Render(modeStr)

	asciiTimer := lipgloss.NewStyle().Margin(1, 0).Render(renderBigTime(m.timeLeft, activeColor, m.width))

	status := "RUNNING"
	if m.paused {
		status = "PAUSED"
	}
	// Make status smaller for very narrow terminals
	var statusStyle lipgloss.Style
	if m.width < 25 {
		statusStyle = lipgloss.NewStyle().Foreground(colorSubtle)
		status = strings.ToLower(status)
	} else {
		statusStyle = lipgloss.NewStyle().Foreground(colorSubtle)
	}
	statusStr := statusStyle.Render(status)

	// Hide help text for screens less than 100% of typical terminal width
	var help string
	if m.width >= 80 {
		helpText := "\n[SPACE] Pause  •  [s] Skip  •  [↑/↓] +/- 1m  •  [q] Quit"
		if m.width < 60 {
			helpText = "\n[SPACE] Pause  •  [s] Skip  •  [↑/↓] +/-1m\n[q] Quit"
		}
		help = styleHelp.Render(helpText)
	}

	return lipgloss.JoinVertical(lipgloss.Center, title, asciiTimer, statusStr, help)
}

func (m model) viewPopup() string {
	var b strings.Builder

	title := "CONFIRMATION"
	if m.width < 30 {
		title = "CONFIRM"
	}
	if m.width < 20 {
		title = "OK?"
	}

	var titleStyle lipgloss.Style
	if m.width < 20 {
		titleStyle = lipgloss.NewStyle().Foreground(colorBlue)
	} else {
		titleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	}
	b.WriteString(titleStyle.Render(title) + "\n\n")

	// Make popup message responsive
	message := m.popupMessage
	if m.width < 40 && strings.Contains(message, "finished!") {
		// Shorten message for narrow terminals
		if strings.Contains(message, "Work session") {
			message = "Start break?"
		} else if strings.Contains(message, "Break finished") {
			message = "Start work?"
		}
	}
	if m.width < 25 {
		// Even shorter for very narrow terminals
		if strings.Contains(message, "break") {
			message = "Break?"
		} else if strings.Contains(message, "work") {
			message = "Work?"
		}
	}
	b.WriteString(message + "\n\n")

	if m.width >= 80 {
		helpText := "[y] Yes  •  [n] No  •  [q] Quit"
		if m.width < 30 {
			helpText = "[y] Yes  •  [n] No\n[q] Quit"
		}
		b.WriteString(styleHelp.Render(helpText))
	}
	return b.String()
}

func main() {
	var help = flag.Bool("help", false, "Show help")
	var h = flag.Bool("h", false, "Show help")
	flag.Parse()

	if *help || *h {
		printUsage()
		return
	}

	args := flag.Args()
	var w, b, s string
	if len(args) > 0 {
		w = args[0]
	}
	if len(args) > 1 {
		b = args[1]
	}
	if len(args) > 2 {
		s = args[2]
	}
	p := tea.NewProgram(initialModel(w, b, s), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}

func printUsage() {
	fmt.Println("Pomodoro CLI - A simple Pomodoro timer for the terminal")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  pomo                    # Interactive mode")
	fmt.Println("  pomo [work] [break] [sessions]  # Quick start")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -h, --help              # Show this help")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  pomo                    # Start interactive setup")
	fmt.Println("  pomo 25m                # 25min work, 5min break, 4 sessions")
	fmt.Println("  pomo 50m 10m            # 50min work, 10min break, 4 sessions")
	fmt.Println("  pomo 45m 15m 6          # 45min work, 15min break, 6 sessions")
	fmt.Println()
	fmt.Println("CONTROLS (in timer mode):")
	fmt.Println("  SPACE                   # Pause/Resume")
	fmt.Println("  s                       # Skip current session")
	fmt.Println("  ↑/↓                     # Adjust time ±1 minute")
	fmt.Println("  ESC                     # Return to setup menu")
	fmt.Println("  q                       # Quit")
	fmt.Println()
	fmt.Println("For more information, see: https://github.com/yourusername/pomodoro-cli")
}
