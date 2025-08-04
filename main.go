package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"os"
	"strings"
)

type LogEntry struct {
	Namespace string      `json:"namespace"`
	Time      string      `json:"time"`
	Priority  string      `json:"priority"`
	Payload   interface{} `json:"payload"`
}

type model struct {
	logs   []string
	offset int
	width  int
	height int
}

type logMsg string

func (m model) Init() bubbletea.Cmd {
	return readLogs
}

func (m model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {
	case logMsg:
		m.logs = append(m.logs, string(msg))
		if len(m.logs) > m.height {
			m.offset = len(m.logs) - m.height
		}
		return m, nil
	case bubbletea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, bubbletea.Quit
		case "up":
			if m.offset > 0 {
				m.offset--
			}
		case "down":
			if m.offset < len(m.logs)-m.height {
				m.offset++
			}
		}
	case bubbletea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	start := m.offset
	end := m.offset + m.height
	if end > len(m.logs) {
		end = len(m.logs)
	}
	for _, log := range m.logs[start:end] {
		b.WriteString(log + "\n")
	}
	return b.String()
}

func readLogs() bubbletea.Msg {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // skip invalid JSON
		}

		// Pretty-print log
		var payloadStr string
		switch v := entry.Payload.(type) {
		case string:
			payloadStr = v
		default:
			b, _ := json.MarshalIndent(v, "", "  ")
			payloadStr = string(b)
		}

		logLine := fmt.Sprintf(
			"%s | %s | %s\n%s",
			lipgloss.NewStyle().Foreground(priorityColor(entry.Priority)).Render(entry.Time),
			entry.Namespace,
			entry.Priority,
			payloadStr,
		)
		return logMsg(logLine)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return logMsg("error reading stdin")
	}

	return nil
}

func priorityColor(p string) lipgloss.Color {
	switch strings.ToLower(p) {
	case "debug":
		return lipgloss.Color("8")
	case "info":
		return lipgloss.Color("10")
	case "warn":
		return lipgloss.Color("11")
	case "error":
		return lipgloss.Color("9")
	default:
		return lipgloss.Color("7")
	}
}

func main() {
	p := bubbletea.NewProgram(model{})
	if err := p.Start(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
