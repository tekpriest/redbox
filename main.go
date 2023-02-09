package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-redis/redis"
)

var entryStyle = lipgloss.NewStyle().Margin(1, 2)

type entry struct {
	key, value string
}

func (e entry) Title() string       { return e.key }
func (e entry) Description() string { return e.value }
func (e entry) FilterValue() string { return e.key }

type model struct {
	list     list.Model
	entries  []entry
	loaded   bool
	quitting bool
}

func connectRedis() *redis.Client {
	redisURL := "127.0.0.1:6379"
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "password",
	})

	if err := rdb.Ping().Err(); err != nil {
		fmt.Printf("there was an error connecting to redis: %v", err)
		os.Exit(1)
	}

	return rdb
}

func readEntries() []entry {
	var entries []entry
	client := connectRedis()

	keys := client.Keys("*").Val()
	for _, v := range keys {
		entries = append(entries, entry{
			key:   v,
			value: client.Get(v).Val(),
		})
	}

	return entries
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		if !m.loaded {
			h, v := entryStyle.GetFrameSize()
			m.list.SetSize(msg.Width-h, msg.Height-v)
			m.loaded = true
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	if m.loaded {
		return entryStyle.Render(m.list.View())
	} else {
		return "loading entries..."
	}
}

func main() {
	entries := readEntries()
	var items []list.Item
	for _, v := range entries {
		items = append(items, v)
	}

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Entries"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
