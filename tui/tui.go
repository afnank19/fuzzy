package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* STYLES */
var border = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#7D56F4"))
var inputStyle = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#bdae93")).PaddingLeft(1).Foreground(lipgloss.Color("#3C3836"))
var selectedItemStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#edf8ff"))

/* STYLES */

type SearchResult struct {
	Word            string
	MinEditDistance int
	Likeness        float32
	Score           int
	Path            string
}

type list struct {
	items  []SearchResult
	cursor int
	height int
	offset int
}

type model struct {
	turing       string
	typedWord    string
	searchResult []SearchResult
	results      list
	windowHeight int
	windowWidth  int
}

func StartTUI() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func initialModel() model {
	//
	sr := initializeFileList()

	return model{
		turing:       "[+][-][*][/]",
		typedWord:    "",
		searchResult: sr,
		windowHeight: 0,
		windowWidth:  0,
		results: list{
			items:  sr,
			cursor: 0,
			height: 13,
			offset: 0,
		},
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update the terminal size when the window is resized
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

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
				RunSearchAlgo(&m.searchResult, nGramedWord)

				// m.searchResult = runLevenshtein(m.searchResult, m.typedWord)
				return m, nil
			}
		case "right", "left", "tab", "enter":

		case "up":
			scrollListUp(&m.results)
		case "down":
			scrollListDown(&m.results)

		default:
			m.typedWord += msg.String()
			// 1. Convert typedWord into nGrams
			nGramedWord := ConvertWordToNGram(m.typedWord)
			// 2. Send nGrams into the algo func
			// Line below passes by value, which creates a deep copy thus increasing memory
			// m.searchResult = runSearchAlgo(m.searchResult, nGramedWord)

			//Implementing everything with pointers to avoid deep copy
			RunSearchAlgo(&m.searchResult, nGramedWord)

			// Levenshtein is really slow on total results > 35K, Use for testing with caution.
			// Trigger the levenshtein here on the typed word
			// m.searchResult = runLevenshtein(m.searchResult, m.typedWord)

			return m, nil
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	var _ = lipgloss.NewStyle().
		Width(m.windowWidth).
		Height(m.windowHeight).
		AlignHorizontal(lipgloss.Left).
		AlignVertical(lipgloss.Bottom).
		Background(lipgloss.Color("#282828")).
		Foreground(lipgloss.Color("#a89984"))
	// var border = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#7D56F4")).Width(int(float32(m.windowWidth)/1.5))
	border = border.Width(int(float32(m.windowWidth) / 1.5))

	inputStyle = inputStyle.Width(m.windowWidth)
	var input string = inputStyle.Render("Query: " + m.typedWord + "|")

	// var displayLimit int = 30
	var dirList string = ""
	// var idx int = len(m.searchResult) - displayLimit
	// for i := len(m.searchResult) - 1; i > len(m.searchResult)-displayLimit; i -= 1 {
	// 	// var likeness string = strconv.FormatFloat(float64(m.searchResult[idx].Likeness), 'f' ,2, 64)
	// 	dirList += " " + m.searchResult[idx].Path + " - " + fmt.Sprintf("%v", m.searchResult[idx].Likeness) + "\n"
	// 	idx++
	// }
	// dirList += selectedItemStyle.Render(" > " + m.searchResult[idx].Path + " - " + fmt.Sprintf("%v", m.searchResult[idx].Likeness) + "\n")

	end := min(m.results.offset+m.results.height, len(m.results.items)) - 1

	for i := end; i >= m.results.offset; i-- {
		cursor := " " // no cursor by default
		currItem := m.results.items[i].Path

		if i == m.results.cursor {
			cursor = ">" // cursor indicator
			// currItem = currItem
		}

		dirList += fmt.Sprintf("%s %s | %f\n", cursor, currItem, m.results.items[i].Likeness)
	}

	// for i := 0; i < len(m.searchResult) && i < 30; i++ {
	// 	// var midEd string = strconv.Itoa(m.searchResult[i].MinEditDistance)
	// 	if  m.searchResult[i].Likeness < 1.0 {
	// 		var likeness string = strconv.FormatFloat(float64(m.searchResult[i].Likeness), 'f' ,2, 64)
	// 		dirList +=" " + m.searchResult[i].Word + " | path: " + m.searchResult[i].Path + " | match: " + likeness  +"\n"
	// 	}
	// }

	return (dirList + "\n\n" + input)
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
