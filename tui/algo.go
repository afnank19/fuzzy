package tui

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func initializeFileList(onlyFolders bool) []SearchResult {
	srs := []SearchResult{}

	// Get the users home directory
	homeDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	count := 0

	DirWalk(homeDir, &srs, &count, onlyFolders)

	return srs
}

func DirWalk(path string, strs *[]SearchResult, count *int, onlyFolders bool) {
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

		homeDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		// In order to achieve similar results to fzf, we can put the Word equal to the trimmed path
		// but for larger datasets. above 300K, performance drop is noteable
		// This achieves results almost as good as fzf
		// But fzf uses a different algorithm that doesn't really perform well if you misspell
		// your query, wherease this algorithm tries to guess what you meant
		// if onlyFolders {
		// 	if info.IsDir() {
		// 		*strs = append(*strs, SearchResult{Word: strings.TrimPrefix(path, homeDir+"/"), MinEditDistance: 999, Likeness: 0.0, Path: strings.TrimPrefix(path, homeDir)})
		// 		*count++
		// 	}
		// } else {
		// 	if !info.IsDir() {
		// 		*strs = append(*strs, SearchResult{Word: strings.TrimPrefix(path, homeDir+"/"), MinEditDistance: 999, Likeness: 0.0, Path: strings.TrimPrefix(path, homeDir)})
		// 		*count++
		// 	}
		// }

		if (onlyFolders && info.IsDir()) || (!onlyFolders && !info.IsDir()) {
			*strs = append(*strs, SearchResult{
				Word:            strings.TrimPrefix(path, homeDir+"/"),
				MinEditDistance: 999,
				Likeness:        0.0,
				Path:            strings.TrimPrefix(path, homeDir),
			})
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

func ConvertWordToNGram(word string) []string {
	var nGrams []string

	//nGrams = strings.SplitAfter(str, "")
	//str = str[0:2]
	for i := 0; i <= len(word)-2; i++ {
		// strSlice :=
		nGrams = append(nGrams, word[i:i+2])
	}

	return nGrams
}

func RunSearchAlgo(sr *[]SearchResult, nGramedWord []string) {

	// 1. Loop over all files
	for i := 0; i < len(*sr); i++ {
		// 2. On each word, run the scoring algo
		(*sr)[i].Score = ScoreWord((*sr)[i].Word, nGramedWord)

		(*sr)[i].Likeness = float64((*sr)[i].Score) * float64((*sr)[i].Score) / float64(len((*sr)[i].Word))
	}

	sort.Slice(*sr, func(i, j int) bool {
		return (*sr)[i].Likeness > (*sr)[j].Likeness
	})
}

func ScoreWord(s string, nGramedWord []string) int {
	var score int = 0

	for i := 0; i < len(nGramedWord); i++ {
		if strings.Contains(s, nGramedWord[i]) {
			score++
		} else {
			score--
		}
	}

	if score < 0 {
		return 0
	}

	return score
}
