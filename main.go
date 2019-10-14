package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/pkliczewski/fileops/compress"
)

const (
	// SourceDir points to folder on PV to write and copy files from for gather data
	SourceDir = "/must-gather/"
	// DestinationFile defines prefix for destination file
	DestinationFile = "must-gather"
	// TerminationLog container termination log
	TerminationLog = "/dev/termination-log"
)

func getValue(name string) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil {
		value = 0
	}
	return value
}

func logResult(message interface{}) {
	f, err := os.OpenFile(TerminationLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println(message)
}

func main() {
	now := time.Now()
	filename := DestinationFile + now.Format("20060102150405") + ".tar.gz"

	mins := getValue("MINUTES")
	numLines := getValue("LINES")

	if mins != 0 && numLines != 0 {
		logResult("Not possible to filter by time and number of lines")
		os.Exit(1)
	}

	// run compress with filter parameters
	err := compress.Compress(filename, SourceDir, numLines, mins)
	if err != nil {
		logResult(err)
		os.Exit(1)
	}
	return
}
