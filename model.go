package kancli

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MoveItemPredicate func(MoveMsg)

type Board struct {
	help       help.Model
	loaded     bool
	Focused    focus
	Cols       []Column
	quitting   bool
	onMoveItem MoveItemPredicate
}

type focus int

func (f focus) Next() focus {
	if f == done {
		return todo
	}
	return f + 1
}

func (f focus) Prev() focus {
	if f == todo {
		return done
	}
	return f - 1
}

const (
	todo focus = iota
	inProgress
	done
)

// NewDefaultBoard creates a new kanban board with To Do, In Progress, and Done
// columns.
func NewDefaultBoard(cols []Column, onMoveItem MoveItemPredicate) *Board {
	help := help.New()
	help.ShowAll = true
	b := &Board{Cols: cols, help: help, onMoveItem: onMoveItem}
	for i, c := range cols {
		if c.Focused() {
			b.Focused = focus(i)
		}
		c.board = b
	}

	b.Cols[todo].List.Title = "To Do"
	b.Cols[inProgress].List.Title = "In Progress"
	b.Cols[done].List.Title = "Done"

	return b
}

func (m *Board) Init() tea.Cmd {
	return nil
}

func (m *Board) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width - margin
		for i := 0; i < len(m.Cols); i++ {
			var res tea.Model
			res, cmd = m.Cols[i].Update(msg)
			m.Cols[i] = res.(Column)
			cmds = append(cmds, cmd)
		}
		m.loaded = true
		return m, tea.Batch(cmds...)
	case MoveMsg:
		cmds = append(cmds, m.Cols[msg.i].Set(APPEND, msg.item))
		m.onMoveItem(msg)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, keys.Left):
			m.Cols[m.Focused].Blur()
			m.Focused = m.Focused.Prev()
			m.Cols[m.Focused].Focus()
		case key.Matches(msg, keys.Right):
			m.Cols[m.Focused].Blur()
			m.Focused = m.Focused.Next()
			m.Cols[m.Focused].Focus()
		}
	}
	res, cmd := m.Cols[m.Focused].Update(msg)
	cmds = append(cmds, cmd)
	if _, ok := res.(Column); ok {
		m.Cols[m.Focused] = res.(Column)
	} else {
		// if it's not a column, switch to the returned model
		return res, tea.Batch(cmds...)
	}
	return m, tea.Batch(cmds...)
}

// Changing to pointer receiver to get back to this model after adding a new task via the form... Otherwise I would need to pass this model along to the form and it becomes highly coupled to the other models.
func (m *Board) View() string {
	if m.quitting {
		return ""
	}
	if !m.loaded {
		return "loading..."
	}
	board := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.Cols[todo].View(),
		m.Cols[inProgress].View(),
		m.Cols[done].View(),
	)
	return lipgloss.JoinVertical(lipgloss.Left, board, m.help.View(keys))
}
