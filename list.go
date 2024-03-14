package vorl

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type replList struct {
	items           []string
	fn              func(string) interface{}
	list            list.Model
	width           int
	height          int
	interactiveMode bool
	executedCommand bool
}

func newList(
	it []string,
	fn func(string) interface{},
	width int,
	height int,
) replList {
	items := make([]list.Item, len(it))
	for i, item := range it {
		items[i] = listItem(item)
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetHeight(1)
	delegate.SetSpacing(0)
	delegate.Styles.SelectedTitle = delegate.Styles.NormalTitle

	height -= 4
	listHeight := min(len(it), height)
	list := list.New(items, delegate, width, listHeight)
	list.SetShowTitle(false)
	list.SetShowFilter(false)
	list.SetFilteringEnabled(false)
	list.DisableQuitKeybindings()
	list.SetShowHelp(false)
	list.SetShowStatusBar(false)

	if len(it) < height {
		list.SetShowPagination(false)
	}

	return replList{
		list:            list,
		interactiveMode: false,
		items:           it,
		fn:              fn,
		width:           width,
		height:          height,
	}
}

func (l replList) View() string {
	return l.list.View()
}

func (l replList) Update(msg tea.Msg) (replList, tea.Cmd) {
	l.executedCommand = false

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetHeight(1)
	delegate.SetSpacing(0)
	l.list.SetDelegate(delegate)

	if !l.interactiveMode {
		l.list.SetShowHelp(false)
		l.list.SetShowStatusBar(false)
		l.list.SetShowFilter(false)
		l.list.SetFilteringEnabled(false)
		delegate.Styles.SelectedTitle = delegate.Styles.NormalTitle
		l.list.SetHeight(min(len(l.items), l.height))
		l.list.SetDelegate(delegate)

		return l, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if l.list.FilterState() == list.Filtering {
				// let the list filter handle the event
				break
			}

			// execute the command associated to the item
			if l.fn != nil {
				cmds = append(cmds, tea.Println(l.list.View()))
				cmds = append(cmds, func() tea.Msg {
					msg := l.fn(string(l.list.SelectedItem().(listItem)))
					if msg == nil {
						msg = CommandResultSimple("")
					}
					return msg
				})
				l.executedCommand = true
			}
		}
	}

	l.list.SetShowFilter(true)
	l.list.SetFilteringEnabled(true)
	l.list.SetShowHelp(true)
	l.list.SetShowStatusBar(true)
	l.list.SetHeight(min(len(l.items)+5, l.height))

	var cmd tea.Cmd
	l.list, cmd = l.list.Update(msg)
	cmds = append(cmds, cmd)

	return l, tea.Batch(cmds...)
}

func (l replList) ExecutedCommand() bool {
	return l.executedCommand
}

func (l replList) RunSelected() (tea.Cmd, bool) {
	if l.fn == nil || !l.InteractiveMode() {
		return nil, false
	}

	return func() tea.Msg {
		return l.fn(string(l.list.SelectedItem().(listItem)))
	}, true
}

func (l *replList) SetInteractiveMode(m bool) {
	// TODO: sacar esto por ahi, solo checkear si hay posibilidad de scroll
	/*
		if l.fn == nil {
			l.selectingMode = false
			return
		}
	*/

	l.interactiveMode = m
}

func (l replList) InteractiveMode() bool {
	return l.interactiveMode
}

type listItem string

func (l listItem) FilterValue() string {
	return string(l)
}

func (l listItem) Title() string {
	return string(l)
}

func (l listItem) Description() string {
	return ""
}
