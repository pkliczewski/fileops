package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pkliczewski/fileops/fileops"

	"github.com/pkliczewski/fileops/compress"
)

func main() {
	var src string
	var dst string

	flag.StringVar(&src, "src", "/home/pkliczewski/go/src/github.com/pkliczewski/fileops/must-gather", "Source directory to be compressed")
	flag.StringVar(&dst, "dst", "file.tar.gz", "Destination archive name")

	var mins int
	var numLines int
	flag.IntVar(&mins, "since", 0, "Filter logs by time in minutes")
	flag.IntVar(&numLines, "lines", 0, "Filter logs by last number of lines")

	if mins != 0 && numLines != 0 {
		fmt.Println("Not possible to filter by time and number of lines")
		os.Exit(1)
	}

	var fn fileops.Check = nil
	if mins > 0 {
		fn = fileops.DatePartial(mins)
	} else if numLines > 0 {
		fn = fileops.LimitPartial(numLines)
	}

	err := compress.Compress(dst, src, fn)
	if err != nil {
		fmt.Println(err)
	}
	return
}
