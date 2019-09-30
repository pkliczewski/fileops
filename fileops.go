package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	filename = "copy.log"
	numLines = 5
	date     = "2019-09-26T12:01:24.150971631Z"
	minutes  = 10
)

func main() {
	// fn := limitPartial(numLines)
	fn := datePartial(minutes)
	text, err := TailFile(filename, fn)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print(text)
	return
}

func limitPartial(numLines int) func([]string, []byte, int) (bool, []string) {
	return func(resultLines []string, line []byte, lines int) (bool, []string) {
		return lines >= numLines-1, append(resultLines, reverse(line))
	}
}

func datePartial(minutes int) func([]string, []byte, int) (bool, []string) {
	// mock current data base on last time stamp from the log
	now, _ := time.Parse(time.RFC3339, date)
	deadline := now.Add(time.Duration(0-minutes) * time.Minute)

	return func(lines []string, line []byte, _ int) (bool, []string) {
		words := strings.Fields(reverse(line))
		lineTime, err := time.Parse(time.RFC3339, words[0])
		if err != nil {
			return true, lines
		}
		diff := deadline.Sub(lineTime)
		return diff >= 0, append(lines, string(line))
	}
}

type check func([]string, []byte, int) (bool, []string)

// TailFile trims end of the file by defined func
func TailFile(filename string, fn check) ([]string, error) {
	if len(filename) == 0 {
		return []string{}, errors.New("you must provide the path to a file")
	}

	file, err := os.Open(filename)
	if err != nil {
		return []string{}, err
	}
	defer file.Close()

	// read file backwards
	lines := 0
	var offset int64 = -1
	var line []byte
	var resultLines []string
	var cond bool
	for {
		// seek to new position in file
		startPos, err := file.Seek(offset, 2)
		if err != nil {
			return []string{}, err
		}

		// make sure start position can never be less than 0
		if startPos == 0 {
			break
		}

		b := make([]byte, 1)
		_, err = file.ReadAt(b, startPos)
		line = append(line, b...)
		if err != nil {
			return []string{}, err
		}

		// ignore if first character being read is a newline
		if offset == int64(-1) && string(b) == "\n" {
			offset--
			line = []byte{}
			continue
		}

		// if the character is a newline add to the number of lines read
		if string(b) == "\n" {
			cond, resultLines = fn(resultLines, line, lines)
			if cond {
				break
			}
			lines++
			line = []byte{}
		}

		offset--
	}

	return reverseLines(resultLines), nil
}

func reverse(a []byte) string {
	for i := len(a)/2 - 1; i >= 0; i-- {
		pos := len(a) - 1 - i
		a[i], a[pos] = a[pos], a[i]
	}
	return string(a)
}

func reverseLines(a []string) []string {
	for i := len(a)/2 - 1; i >= 0; i-- {
		pos := len(a) - 1 - i
		a[i], a[pos] = a[pos], a[i]
	}
	return a
}
