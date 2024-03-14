package vorl

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type replInput struct {
	textInput       textinput.Model
	execFn          func(string) tea.Cmd
	suggestFn       func(string) []string
	history         []string
	executedCommand bool
}

func newInput(prompt string, execFn func(string) tea.Cmd, suggestFn func(string) []string) replInput {
	textInput := textinput.New()
	textInput.Prompt = prompt + " "
	textInput.ShowSuggestions = true
	textInput.Focus()

	return replInput{
		textInput: textInput,
		execFn:    execFn,
		suggestFn: suggestFn,
	}
}

func (ri replInput) Update(msg tea.Msg) (replInput, tea.Cmd) {
	ri.executedCommand = false
	var cmds []tea.Cmd

	input := ri.textInput.Value()
	if msg, ok := msg.(tea.KeyMsg); ok && msg.Type == tea.KeyEnter {
		ri.textInput.SetValue("")

		if input != "" {
			cmds = append(cmds, tea.Printf("%s%s", ri.textInput.Prompt, input))

			if ri.execFn != nil {
				cmds = append(
					cmds,
					ri.execFn(input),
				)
				ri.executedCommand = true
				ri.history = append(ri.history, input)
			}
		}
	}

	suggestions := ri.suggestFn(input)
	suggestions = append(ri.history, suggestions...)
	ri.textInput.SetSuggestions(suggestions)

	var cmd tea.Cmd
	ri.textInput, cmd = ri.textInput.Update(msg)
	cmds = append(cmds, cmd)

	return ri, tea.Batch(cmds...)
}

func (ri replInput) ExecutedCommand() bool {
	return ri.executedCommand
}

func (ri replInput) Value() string {
	return ri.textInput.Value()
}

func (ri replInput) View() string {
	return ri.textInput.View()
}
