package vorl

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type replState int

const (
	replStateReadingInput replState = iota
	replStateReadingInputAndList
	replStateReadingInputAndTable
	replStateExecutingCommand
	replStateListInteraction
	replStateTableInteraction
)

type Interpreter interface {
	Exec(command string) (interface{}, error)

	Suggest(partialInput string) []string
}

type REPL struct {
	model model
}

func NewREPL(interpreter Interpreter) *REPL {
	return &REPL{
		model: initialModel(interpreter),
	}
}

func (r *REPL) Run() error {
	p := tea.NewProgram(r.model)

	_, err := p.Run()
	return err
}

type model struct {
	interpreter Interpreter

	textInput replInput

	state replState

	listResult *replList

	tableResult *replTable

	spinner spinner.Model

	height int
	width  int
}

func initialModel(interpreter Interpreter) model {
	execFn := func(cmd string) tea.Cmd {
		return func() tea.Msg {
			result, err := interpreter.Exec(cmd)
			if err != nil {
				return commandError(err)
			}

			return result
		}
	}

	input := newInput("vor >", execFn, interpreter.Suggest)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		interpreter: interpreter,
		textInput:   input,
		state:       replStateReadingInput,
		spinner:     sp,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlD:
			return m, tea.Quit

		case tea.KeyCtrlUp, tea.KeyCtrlK:
			if m.listResult != nil {
				m.state = replStateListInteraction
				m.listResult.SetInteractiveMode(true)
			}

			if m.tableResult != nil {
				m.state = replStateTableInteraction
				m.tableResult.SetInteractiveMode(true)
			}

		case tea.KeyCtrlDown, tea.KeyCtrlJ:
			if m.listResult != nil {
				m.state = replStateReadingInputAndList
				m.listResult.SetInteractiveMode(false)

				newList, cmd := m.listResult.Update(msg)
				m.listResult = &newList
				cmds = append(cmds, cmd)
			}

			if m.tableResult != nil {
				m.state = replStateReadingInputAndTable
				m.tableResult.SetInteractiveMode(false)

				newTable, cmd := m.tableResult.Update(msg)
				m.tableResult = &newTable
				cmds = append(cmds, cmd)
			}

		case tea.KeyEnter:
			if m.state == replStateReadingInputAndList && m.textInput.Value() != "" {
				// if the list was not used, print it and remove it
				cmd := tea.Println(m.listResult.View())
				cmds = append(cmds, cmd)
				m.listResult = nil
				m.state = replStateReadingInput
			}

			if m.state == replStateReadingInputAndTable && m.textInput.Value() != "" {
				// if the table was not used, print it and remove it
				cmd := tea.Println(m.tableResult.View())
				cmds = append(cmds, cmd)
				m.tableResult = nil
				m.state = replStateReadingInput
			}
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	case commandError:
		cmds = append(cmds, tea.Printf("ERROR: %v", msg))
		m.listResult = nil
		m.tableResult = nil
		m.state = replStateReadingInput

	case CommandResultSimple:
		cmds = append(cmds, tea.Printf("%+v", msg))
		m.listResult = nil
		m.tableResult = nil
		m.state = replStateReadingInput

	case CommandResultList:
		l := newList(msg.List, msg.OnSelect, m.width, m.height)
		m.listResult = &l
		m.tableResult = nil
		m.state = replStateReadingInputAndList

	case CommandResultTable:
		table := newTable(msg.Table, msg.OnSelect, m.width, m.height)
		m.listResult = nil
		m.tableResult = &table
		m.state = replStateReadingInputAndTable

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch m.state {
	case replStateReadingInput,
		replStateReadingInputAndList,
		replStateReadingInputAndTable:

		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

		if m.textInput.ExecutedCommand() {
			m.state = replStateExecutingCommand
		}

	case replStateListInteraction:
		newList, cmd := m.listResult.Update(msg)
		m.listResult = &newList
		cmds = append(cmds, cmd)

		if m.listResult.ExecutedCommand() {
			m.state = replStateExecutingCommand
			m.listResult = nil
		}

	case replStateTableInteraction:
		newTable, cmd := m.tableResult.Update(msg)
		m.tableResult = &newTable
		cmds = append(cmds, cmd)

		if m.tableResult.ExecutedCommand() {
			m.state = replStateExecutingCommand
			m.tableResult = nil
		}

	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	view := ""

	if m.listResult != nil {
		view += m.listResult.View() + "\n"
	}

	if m.tableResult != nil {
		view += m.tableResult.View() + "\n"
	}

	if m.state == replStateExecutingCommand {
		view += fmt.Sprintf("%s executing...\n", m.spinner.View())
	} else {
		view += m.textInput.View() + "\n"
	}
	return view
}

func (m model) executeCommand(cmd string) tea.Cmd {
	return func() tea.Msg {
		result, err := m.interpreter.Exec(cmd)
		if err != nil {
			return commandError(err)
		}

		return result
	}
}

type commandError error

type CommandResultSimple string

type CommandResultList struct {
	List     []string
	OnSelect func(selected string) interface{}
}

type CommandResultTable struct {
	Table    [][]string
	OnSelect func(selected []string) interface{}
}
