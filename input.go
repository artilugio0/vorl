package vorl

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type replInputState int

const (
	replInputStateReadingInput replInputState = iota
	replInputStateReverseSearch
)

type replInput struct {
	historyIndex    int
	textInput       textinput.Model
	execFn          func(string) tea.Cmd
	suggestFn       func(string) []string
	history         []string
	executedCommand bool

	state                replInputState
	reverseSearchInput   string
	reverseSearchResults []string
	reverseSearchIndex   int
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

	switch ri.state {
	case replInputStateReadingInput:
		var cmd tea.Cmd
		ri, cmd = ri.readingInputUpdate(msg)
		cmds = append(cmds, cmd)

	case replInputStateReverseSearch:
		var cmd tea.Cmd
		ri, cmd = ri.reverseSearchUpdate(msg)
		cmds = append(cmds, cmd)
	}

	return ri, tea.Batch(cmds...)
}
func (ri replInput) readingInputUpdate(msg tea.Msg) (replInput, tea.Cmd) {
	var cmds []tea.Cmd

	input := ri.textInput.Value()
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyEnter:
			ri.textInput.SetValue("")
			ri.historyIndex = 0

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

		case tea.KeyCtrlC:
			ri.textInput.SetValue("")
			ri.historyIndex = 0

		case tea.KeyUp:
			ri.historyIndex = min(ri.historyIndex+1, len(ri.history))

		case tea.KeyDown:
			ri.historyIndex = max(ri.historyIndex-1, 0)

		case tea.KeyCtrlR:
			ri.state = replInputStateReverseSearch

		default:
			ri.historyIndex = 0
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

func (ri replInput) reverseSearchUpdate(msg tea.Msg) (replInput, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyEnter:
			newInput := ""
			if len(ri.reverseSearchResults) > 0 {
				newInput = ri.reverseSearchResults[ri.reverseSearchIndex]
			}
			ri.textInput.SetValue(newInput)
			ri.textInput.SetCursor(len(newInput))

			ri.reverseSearchInput = ""
			ri.state = replInputStateReadingInput
			ri.reverseSearchResults = []string{}

		case tea.KeyCtrlC:
			ri.reverseSearchInput = ""
			ri.state = replInputStateReadingInput
			ri.reverseSearchResults = []string{}

		case tea.KeyCtrlR:
			ri.reverseSearchIndex = min(ri.reverseSearchIndex+1, len(ri.reverseSearchResults)-1)

		case tea.KeyBackspace:
			if len(ri.reverseSearchInput) == 0 {
				break
			}

			ri.reverseSearchInput = ri.reverseSearchInput[:len(ri.reverseSearchInput)-1]

			if len(ri.reverseSearchInput) == 0 {
				ri.reverseSearchResults = []string{}
				break
			}

			for _, histInput := range ri.history {
				if strings.Contains(histInput, ri.reverseSearchInput) {
					ri.reverseSearchResults = append(ri.reverseSearchResults, histInput)
				}
			}

			slices.Reverse(ri.reverseSearchResults)

		default:
			if len(msg.String()) > 1 {
				break
			}

			ri.reverseSearchIndex = 0
			ri.reverseSearchInput += msg.String()
			ri.reverseSearchResults = []string{}

			for _, histInput := range ri.history {
				if strings.Contains(histInput, ri.reverseSearchInput) {
					ri.reverseSearchResults = append(ri.reverseSearchResults, histInput)
				}
			}

			slices.Reverse(ri.reverseSearchResults)
		}
	}

	return ri, nil
}

func (ri replInput) ExecutedCommand() bool {
	return ri.executedCommand
}

func (ri replInput) Value() string {
	return ri.textInput.Value()
}

func (ri replInput) View() string {
	switch ri.state {
	case replInputStateReverseSearch:
		searchResult := ""
		if len(ri.reverseSearchResults) > 0 {
			searchResult = ri.reverseSearchResults[ri.reverseSearchIndex]
		}
		return "rs: '" + ri.reverseSearchInput + "' " + ri.textInput.Prompt + " " + searchResult

	case replInputStateReadingInput:
		return ri.textInput.View()

	default:
		return ri.textInput.View()
	}
}
