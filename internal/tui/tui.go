package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/maskedsyntax/jvman/internal/config"
	"github.com/maskedsyntax/jvman/internal/registry"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginLeft(2)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginLeft(2)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	currentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("78")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginLeft(2).
			MarginTop(1)
)

type item struct {
	name     string
	vendor   string
	isCurrent bool
}

func (i item) Title() string {
	if i.isCurrent {
		return currentStyle.Render("* " + i.name)
	}
	return "  " + i.name
}

func (i item) Description() string {
	return i.vendor
}

func (i item) FilterValue() string {
	return i.name
}

type keyMap struct {
	Switch key.Binding
	Remove key.Binding
	Quit   key.Binding
}

var keys = keyMap{
	Switch: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "set as global"),
	),
	Remove: key.NewBinding(
		key.WithKeys("d", "delete"),
		key.WithHelp("d", "remove"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type model struct {
	list     list.Model
	cfg      *config.Config
	reg      *registry.Registry
	status   string
	quitting bool
}

func initialModel(cfg *config.Config, reg *registry.Registry) model {
	items := buildItemList(cfg, reg)

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = selectedStyle
	delegate.Styles.NormalTitle = normalStyle

	l := list.New(items, delegate, 60, 20)
	l.Title = "jvman - Installed Java Versions"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return model{
		list:   l,
		cfg:    cfg,
		reg:    reg,
		status: "",
	}
}

func buildItemList(cfg *config.Config, reg *registry.Registry) []list.Item {
	installed := reg.List()
	items := make([]list.Item, 0, len(installed))

	for name, jvm := range installed {
		items = append(items, item{
			name:      name,
			vendor:    jvm.Vendor,
			isCurrent: name == cfg.Global,
		})
	}

	return items
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, keys.Switch):
			if i, ok := m.list.SelectedItem().(item); ok {
				if err := m.reg.SetGlobal(i.name); err != nil {
					m.status = fmt.Sprintf("Error: %v", err)
				} else {
					m.status = fmt.Sprintf("Switched to %s", i.name)
					m.cfg.Global = i.name
					m.list.SetItems(buildItemList(m.cfg, m.reg))
				}
			}
			return m, nil

		case key.Matches(msg, keys.Remove):
			if i, ok := m.list.SelectedItem().(item); ok {
				if i.isCurrent {
					m.status = "Cannot remove the current global version"
				} else {
					if err := m.reg.Remove(i.name); err != nil {
						m.status = fmt.Sprintf("Error: %v", err)
					} else {
						m.status = fmt.Sprintf("Removed %s", i.name)
						m.list.SetItems(buildItemList(m.cfg, m.reg))
					}
				}
			}
			return m, nil
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

	var b strings.Builder
	b.WriteString(m.list.View())

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(statusStyle.Render(m.status))
	}

	help := helpStyle.Render("enter: set global | d: remove | /: filter | q: quit")
	b.WriteString("\n")
	b.WriteString(help)

	return b.String()
}

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reg := registry.New(cfg)

	if len(reg.List()) == 0 {
		return fmt.Errorf("no Java versions installed. Run 'jvman install <version>' first")
	}

	p := tea.NewProgram(initialModel(cfg, reg), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
