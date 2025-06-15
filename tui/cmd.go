package tui

import (
	"os"
	"os/exec"
)

// tools default value should be "code"!
// doesnt work for terminal based editors yet
func OpenFiles(queuedFiles []string, tool string) {
	for _, filePath := range queuedFiles {
		cmd := exec.Command(tool, filePath)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	}
}
