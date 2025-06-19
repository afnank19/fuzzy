package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* STYLES */
var inputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#e0def4"))
var selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#edf8ff"))
var tabbedItem = lipgloss.NewStyle().Foreground(lipgloss.Color("#ebbcba"))
var cursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#eb6f92"))

/* STYLES */

type SearchResult struct {
	Word            string
	MinEditDistance int
	Likeness        float64
	Score           int
	Path            string
}

type list struct {
	items  []SearchResult
	cursor int
	height int
	offset int
}

type Model struct {
	turing       string
	typedWord    string
	tool         string
	windowHeight int
	windowWidth  int
	results      list
	queuedFiles  []string
	fl           *FileList
	// onlyFolders  bool
}

type FileList struct {
	QueuedFiles []string
	CanOpen     bool
}

func StartTUI(tool string, qf *FileList, oF bool) {
	p := tea.NewProgram(initialModel(tool, qf, oF), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func initialModel(tool string, qf *FileList, oF bool) Model {
	sr := initializeFileList(oF)

	return Model{
		turing:       "[+][-][*][/]",
		typedWord:    "",
		tool:         tool,
		windowHeight: 0,
		windowWidth:  0,
		results: list{
			items:  sr,
			cursor: 0,
			height: 13,
			offset: 0,
		},
		queuedFiles: []string{},
		fl:          qf,
	}
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update the terminal size when the window is resized
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.results.height = msg.Height - 4

	// Is it a key press?
	case tea.KeyMsg:
		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "backspace":
			//Removes the last character of the string
			if len(m.typedWord) > 0 {
				m.typedWord = m.typedWord[:len(m.typedWord)-1]
				nGramedWord := ConvertWordToNGram(m.typedWord)
				RunSearchAlgo(&m.results.items, nGramedWord)

				// m.searchResult = runLevenshtein(m.searchResult, m.typedWord)
				return m, nil
			}
		case "right", "left":
			// do nothing
		case "up":
			scrollListUp(&m.results)
		case "down":
			scrollListDown(&m.results)
		case "enter":
			m.queuedFiles = append(m.queuedFiles, m.results.items[m.results.cursor].Word)
			m.fl.QueuedFiles = m.queuedFiles
			m.fl.CanOpen = true
			// OpenFiles(m.queuedFiles, m.tool)
			return m, tea.Quit
		case "tab": // this needs to be a toggle
			m.queuedFiles = append(m.queuedFiles, m.results.items[m.results.cursor].Word)
			m.fl.QueuedFiles = m.queuedFiles
		default:
			m.typedWord += msg.String()
			// 1. Convert typedWord into nGrams
			nGramedWord := ConvertWordToNGram(m.typedWord)
			// 2. Send nGrams into the algo func
			// Line below passes by value, which creates a deep copy thus increasing memory
			// m.searchResult = runSearchAlgo(m.searchResult, nGramedWord)

			//Implementing everything with pointers to avoid deep copy
			RunSearchAlgo(&m.results.items, nGramedWord)

			// Levenshtein is really slow on total results > 35K, Use for testing with caution.
			// Trigger the levenshtein here on the typed word
			// m.searchResult = runLevenshtein(m.searchResult, m.typedWord)

			return m, nil
		}
	}

	// Return the updated Model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m Model) View() string {
	var view = lipgloss.NewStyle().Width(m.windowWidth).Height(m.windowHeight).AlignHorizontal(lipgloss.Left).AlignVertical(lipgloss.Bottom)
	var rule = lipgloss.NewStyle().Width(m.windowWidth).Border(lipgloss.NormalBorder(), true, false, false, false).BorderForeground(lipgloss.Color("#9ccfd8"))

	// inputStyle = inputStyle.Width(m.windowWidth)
	var input string = inputStyle.Render(": " + m.typedWord + "|")

	var dirList string = ""
	end := min(m.results.offset+m.results.height, len(m.results.items)) - 1

	for i := end; i >= m.results.offset; i-- {
		var item string

		cursor := " " // no cursor by default
		currItem := m.results.items[i].Word
		item = cursor + " " + currItem

		if i == m.results.cursor {
			cursor = cursorStyle.Render(">") // cursor indicator
			// currItem = currItem
			item = cursor + " " + selectedItemStyle.Render(currItem)
		}

		for _, qFiles := range m.queuedFiles {
			if qFiles == currItem {
				item = cursor + " " + tabbedItem.Render(currItem)
			}
		}

		dirList += fmt.Sprintf("%s\n", item)
	}

	currTotalFiles := fmt.Sprintf("%d items", len(m.results.items))

	return view.Render(dirList + rule.Render(currTotalFiles) + "\n" + input)
}

func scrollListDown(list *list) {
	if list.cursor > 0 {
		list.cursor--
		if list.cursor < list.offset {
			list.offset--
		}
	}
}

func scrollListUp(list *list) {
	if list.cursor < len(list.items)-1 {
		list.cursor++
		if list.cursor >= list.offset+list.height {
			list.offset++
		}
	}
}
