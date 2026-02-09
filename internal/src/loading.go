package src

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	padding  = 2
	maxWidth = 80
)

const (
	progressColorStart = "#0055ff"
	progressColorEnd   = "#00d4ff"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

type tickMsg time.Time

type LoadingModel struct {
	Progress    progress.Model
	Description string
}

func NewLoadingModel(description string) LoadingModel {
	return LoadingModel{
		Progress: progress.New(
			progress.WithGradient(progressColorStart, progressColorEnd),
		),
		Description: description,
	}
}

func (m LoadingModel) Init() tea.Cmd {
	return tickCmd()
}

func (m LoadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.Progress.Width = msg.Width - padding*2 - 4
		if m.Progress.Width > maxWidth {
			m.Progress.Width = maxWidth
		}
		return m, nil

	case tickMsg:
		if m.Progress.Percent() == 1.0 {
			return m, tea.Quit
		}
		cmd := m.Progress.IncrPercent(0.25)
		return m, tea.Batch(tickCmd(), cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.Progress.Update(msg)
		m.Progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m LoadingModel) View() string {
	pad := strings.Repeat(" ", padding)
	out := "\n"
	if m.Description != "" {
		out += pad + m.Description + "\n\n"
	}
	out += pad + m.Progress.View() + "\n\n" +
		pad + helpStyle("Press any key to quit")
	return out
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
