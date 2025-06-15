package main

import (
	"flag"

	"github.com/afnank19/fuzzy/tui"
)

func main() {
	tool := flag.String("t", "code", "App you want to open the file\n - (must be able to run through a terminal command)\n - Example: zed")
	flag.Parse()

	var queuedFiles tui.FileList // declared here so it maintains state
	tui.StartTUI(*tool, &queuedFiles)
	tui.OpenFiles(queuedFiles.QueuedFiles, *tool)
}
