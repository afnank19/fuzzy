package main

import (
	"distance/levenshtein/internal"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	PaddingTop(2).
	PaddingLeft(4)

var border = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#7D56F4"))
var inputStyle = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#bdae93")).PaddingLeft(1).Foreground(lipgloss.Color("#3C3836"))
var listStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
var selectedItemStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#edf8ff"))

// type SearchResult struct {
// 	Word            string
// 	MinEditDistance int
// 	Likeness        float32
// 	Score           int
// 	Path            string
// }

type model struct {
	turing       string
	typedWord    string
	searchResult []internal.SearchResult
	windowHeight int
	windowWidth  int
}

func main() {
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
				nGramedWord := convertWordToNGram(m.typedWord)
				runSearchAlgo(&m.searchResult, nGramedWord)

				// m.searchResult = runLevenshtein(m.searchResult, m.typedWord)
				return m, nil
			}
		case "right", "down", "left", "up", "tab", "enter":

		default:
			m.typedWord += msg.String()
			// 1. Convert typedWord into nGrams
			nGramedWord := convertWordToNGram(m.typedWord)
			// 2. Send nGrams into the algo func
			// Line below passes by value, which creates a deep copy thus increasing memory
			// m.searchResult = runSearchAlgo(m.searchResult, nGramedWord)

			//Implementing everything with pointers to avoid deep copy
			runSearchAlgo(&m.searchResult, nGramedWord)

			// Levenshtein is really slow on total files > 35K, Use for testing with caution.
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
	var terminalCenter = lipgloss.NewStyle().
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

	var displayLimit int = 30
	var dirList string = ""
	var idx int = len(m.searchResult) - displayLimit
	for i := len(m.searchResult) - 1; i > len(m.searchResult)-displayLimit; i -= 1 {
		// var likeness string = strconv.FormatFloat(float64(m.searchResult[idx].Likeness), 'f' ,2, 64)
		dirList += " " + m.searchResult[idx].Path + "\n"
		idx++
	}
	dirList += selectedItemStyle.Render(" > " + m.searchResult[idx].Path + "\n")

	// for i := 0; i < len(m.searchResult) && i < 30; i++ {
	// 	// var midEd string = strconv.Itoa(m.searchResult[i].MinEditDistance)
	// 	if  m.searchResult[i].Likeness < 1.0 {
	// 		var likeness string = strconv.FormatFloat(float64(m.searchResult[i].Likeness), 'f' ,2, 64)
	// 		dirList +=" " + m.searchResult[i].Word + " | path: " + m.searchResult[i].Path + " | match: " + likeness  +"\n"
	// 	}
	// }

	return terminalCenter.Render(dirList + "\n\n" + input)
}

func initializeFileList() []internal.SearchResult {
	srs := []internal.SearchResult{}

	// Get the users home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	count := 0

	// irWalk(homeDir, &srs, &count)

	internal.DirWalk(homeDir, &srs, &count)

	fmt.Print("Total files (non-dir) read: ")
	fmt.Println(count)

	return srs
}

func DirWalk(path string, strs *[]internal.SearchResult, count *int) {
	// This can either take the path variable or use "./"
	// To be modified according to use case
	err := filepath.WalkDir(path, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		// Enabling this brings the total file count on my system from 450K to 350K (roughly)
		// Usually skip because these folders don't contain usefull files
		// but can be used to stress test
		if info.IsDir() && (info.Name() == ".git") {
			return filepath.SkipDir
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		// In order to achieve similar results to fzf, we can put the Word equal to the trimmed path
		// but for larger datasets. above 300K, performance drop is noteable
		// This achieves results almost as good as fzf
		// But fzf uses a different algorithm that doesn't really perform well if you misspell
		// your query, wherease this algorithm tries to guess what you meant
		if !info.IsDir() {
			*strs = append(*strs, internal.SearchResult{Word: info.Name(), MinEditDistance: 999, Likeness: 0.0, Path: strings.TrimPrefix(path, homeDir)})
			*count++
		}

		// Code below skips hidden folders, enabling brought my file count to 13K from 350K
		if info.Name()[0] == '.' && info.IsDir() {
			return filepath.SkipDir
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}

func isHiddenDir(path string) bool {
	if path[0] == '.' {
		return true
	}
	return false // Check if it starts with a dot
}

func runLevenshtein(srs []internal.SearchResult, input string) []internal.SearchResult {
	for i := 0; i < len(srs); i++ {
		srs[i] = levenshtein(input, srs[i])
	}

	//Sort by first letter matching the input's first letter

	// Sorts the words by least min edit distance
	// sort.Slice(srs, func(i, j int) bool {
	// 	return srs[i].MinEditDistance > srs[j].MinEditDistance
	// })

	if len(input) > 0 {
		sort.Slice(srs, func(i, j int) bool {
			return srs[i].Likeness < srs[j].Likeness
		})
	}
	return srs
}

func levenshtein(str_1 string, sr internal.SearchResult) internal.SearchResult {
	var str_2 string = sr.Word //Get it from some dict

	var inputLength = len(str_1) + 1
	var cmprsnLength = len(str_2) + 1 //Length of the string being compared with

	leven := make([][]int, inputLength)
	for i := range leven {
		leven[i] = make([]int, cmprsnLength)
	}

	//Initializing the array with default values
	for i := 0; i < inputLength; i++ {
		leven[i][0] = i
	}
	for i := 0; i < cmprsnLength; i++ {
		leven[0][i] = i
	}

	for i := 1; i < inputLength; i++ {
		for j := 1; j < cmprsnLength; j++ {
			var calculated_cost = minOfThree(float64(leven[i-1][j-1]), float64(leven[i-1][j]), float64(leven[i][j-1]))

			if str_1[i-1] != str_2[j-1] {
				calculated_cost += 1
			}

			leven[i][j] = int(calculated_cost)
		}
	}

	sr.MinEditDistance = leven[inputLength-1][cmprsnLength-1]
	sr.Likeness = float32(leven[inputLength-1][cmprsnLength-1]) / float32(len(str_2))

	return sr
}

func minOfThree(a, b, c float64) float64 {
	return math.Min(a, math.Min(b, c))
}

func convertWordToNGram(word string) []string {
	var nGrams []string

	//nGrams = strings.SplitAfter(str, "")
	//str = str[0:2]
	for i := 0; i <= len(word)-2; i++ {
		// strSlice :=
		nGrams = append(nGrams, word[i:i+2])
	}

	return nGrams
}

func runSearchAlgo(sr *[]internal.SearchResult, nGramedWord []string) {

	// 1. Loop over all files
	for i := 0; i < len(*sr); i++ {
		// 2. On each word, run the scoring algo
		(*sr)[i].Score = scoreWord((*sr)[i].Word, nGramedWord)

		(*sr)[i].Likeness = float32((*sr)[i].Score) / float32(len((*sr)[i].Path))
	}

	sort.Slice(*sr, func(i, j int) bool {
		return (*sr)[i].Likeness < (*sr)[j].Likeness
	})
}

func scoreWord(s string, nGramedWord []string) int {
	var score int = 0

	for i := 0; i < len(nGramedWord); i++ {
		if strings.Contains(s, nGramedWord[i]) {
			score++
		}
	}

	return score
}
