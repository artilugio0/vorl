package vorl

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type replTable struct {
	table           table.Model
	interactiveMode bool
	execFn          func([]string) interface{}
	executedCommand bool
}

func newTable(
	rows [][]string,
	execFn func([]string) interface{},
	width int,
	height int,
) replTable {
	tableRows := make([]table.Row, len(rows)-1)
	for i, r := range rows[1:] {
		tableRows[i] = r
	}

	tableColumns := make([]table.Column, len(rows[0]))
	for i, r := range rows[0] {
		tableColumns[i] = table.Column{
			Title: r,
			Width: width / len(rows[0]),
		}
	}

	t := table.New(
		table.WithColumns(tableColumns),
		table.WithRows(tableRows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows)-1, height-6)),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	s.Selected = s.Cell.Copy()
	s.Selected.Padding(0)
	s.Selected.Margin(0)

	t.SetStyles(s)

	return replTable{
		table:  t,
		execFn: execFn,
	}
}

func (rt replTable) View() string {
	return rt.table.View()
}

func (rt replTable) Update(msg tea.Msg) (replTable, tea.Cmd) {
	rt.executedCommand = false

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	if !rt.interactiveMode {
		s.Selected = s.Cell.Copy()
		s.Selected.Padding(0)
		s.Selected.Margin(0)
		rt.table.SetStyles(s)
		return rt, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if rt.execFn != nil && len(rt.table.Rows()) > 0 {
				row := rt.table.SelectedRow()
				cmds = append(cmds, tea.Println(rt.table.View()))
				cmds = append(cmds, func() tea.Msg {
					msg := rt.execFn(row)
					if msg == nil {
						msg = CommandResultSimple("")
					}
					return msg
				})
				rt.executedCommand = true
			}
		}
	}

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	rt.table.SetStyles(s)

	var cmd tea.Cmd
	rt.table, cmd = rt.table.Update(msg)
	cmds = append(cmds, cmd)

	return rt, tea.Batch(cmds...)
}

func (rt *replTable) SetInteractiveMode(enabled bool) {
	rt.interactiveMode = enabled
}

func (rt replTable) ExecutedCommand() bool {
	return rt.executedCommand
}
