package main

import (
	"flag"

	"github.com/afnank19/fuzzy/tui"
)

func main() {
	tool := flag.String("t", "code", "App you want to open the file\n - (must be able to run through a terminal command)\n - Example: zed")
	f := flag.Bool("f", false, "This will consider only folders on your system, useful for opening full projects")
	folder := flag.Bool("folders", false, "This will consider only folders on your system, useful for opening full projects")
	flag.Parse()

	onlyFolders := *f || *folder
	var queuedFiles tui.FileList // declared here so it maintains state
	tui.StartTUI(*tool, &queuedFiles, onlyFolders)

	if queuedFiles.CanOpen {
		tui.OpenFiles(queuedFiles.QueuedFiles, *tool)
	}
}
