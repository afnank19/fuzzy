package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type SearchResult struct {
	Word            string
	MinEditDistance int
	Likeness        float32
	Score           int
	Path            string
}

func DirWalk(path string, strs *[]SearchResult, count *int) {
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
			*strs = append(*strs, SearchResult{Word: info.Name(), MinEditDistance: 999, Likeness: 0.0, Path: strings.TrimPrefix(path, homeDir)})
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