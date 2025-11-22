package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"bl/tindeq"
)

func main() {
	logfilePath := os.Getenv("LOG_FILE")
	if logfilePath == "" {
		logfilePath = "log.txt"
	}
	if _, err := tea.LogToFile(logfilePath, "tindeq"); err != nil {
		log.Fatal(err)
	}

	p := tea.NewProgram(model{status: "starting"})
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	device *tindeq.Device
	status string
}

func (m model) Init() tea.Cmd {
	return scan
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "b":
			if m.device != nil {
				err := m.device.SampleBattery()
				if err != nil {
					log.Fatal(err)
				}
			}
			return m, nil
		case "m":
			if m.device != nil {
				err := m.device.StartMeasure()
				if err != nil {
					log.Fatal(err)
				}
			}
			return m, nil
		case "s":
			if m.device != nil {
				err := m.device.StopMeasure()
				if err != nil {
					log.Fatal(err)
				}
			}
			return m, nil
		case "p":
			if m.device != nil {
				err := m.device.Shutdown()
				if err != nil {
					log.Fatal(err)
				}
				return m, tea.Quit
			}
			return m, nil
		case "q", "ctrl+c":
			if m.device != nil {
				err := m.device.Close()
				if err != nil {
					log.Fatal(err)
				}
			}
			return m, tea.Quit

		}

	case scanResultMsg:
		if msg.err != nil {
			m.status = msg.err.Error()
			return m, nil
		}

		m.status = "connected"
		m.device = msg.device

		return m, m.tick
	case tickMsg:
		return m, m.tick
	}
	return m, nil
}

func (m model) View() string {
	if m.device != nil {
		latest := m.device.GetLatestMeasurement()
		return fmt.Sprintf("%.2f (%d)\n\nBat: %s - %s", latest.Weight, latest.Millis, m.device.GetBattery(), m.status)
	}

	return m.status
}

type scanResultMsg struct {
	device *tindeq.Device
	err    error
}

func scan() tea.Msg {
	log.Printf("scanning\n")
	d, err := tindeq.Scan()
	return scanResultMsg{device: d, err: err}
}

type tickMsg struct{}

func (m model) tick() tea.Msg {
	if m.device == nil {
		return tickMsg{}
	}
	<-m.device.GetUpdateCh()

	return tickMsg{}
}
