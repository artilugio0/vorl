package vorl

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type replInput struct {
	historyIndex    int
	textInput       textinput.Model
	execFn          func(string) tea.Cmd
	suggestFn       func(string) []string
	history         []string
	executedCommand bool
}

func newInput(
	prompt string,
	execFn func(string) tea.Cmd,
	suggestFn func(string) []string,
	initialHistory []string,
) replInput {

	textInput := textinput.New()
	textInput.Prompt = prompt + " "
	textInput.ShowSuggestions = true
	textInput.Focus()

	return replInput{
		textInput: textInput,
		execFn:    execFn,
		suggestFn: suggestFn,
		history:   initialHistory,
	}
}

func (ri replInput) Update(msg tea.Msg) (replInput, tea.Cmd) {
	ri.executedCommand = false
	var cmds []tea.Cmd

	input := ri.textInput.Value()
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyEnter:
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

		case tea.KeyUp:
			ri.historyIndex = min(ri.historyIndex+1, len(ri.history))

		case tea.KeyDown:
			ri.historyIndex = max(ri.historyIndex-1, 0)
		}
	}

	if ri.historyIndex != 0 {
		ri.textInput.SetValue(ri.history[len(ri.history)-ri.historyIndex])
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
