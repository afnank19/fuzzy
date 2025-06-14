package tui

import (
	"os/exec"
)

func openFiles(queuedFiles []string) {
	for _, filePath := range queuedFiles {
		cmd := exec.Command("code", filePath)

		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	}
}
