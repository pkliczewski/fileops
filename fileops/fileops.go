package fileops

import (
	"bufio"
	"errors"
	"os"
	"strings"
	"time"
)

var (
	date = "2019-09-26T12:01:24.150971631Z"
)

// LimitPartial allows to provide line limit and return condition function
func LimitPartial(numLines int) func([]string, []byte, int) (bool, []string, error) {
	return func(resultLines []string, line []byte, lines int) (bool, []string, error) {
		return lines >= numLines-1, append(resultLines, reverse(line)), nil
	}
}

// DatePartial allows to provide date and return condition function
func DatePartial(minutes int) func([]string, []byte, int) (bool, []string, error) {
	// mock current data base on last time stamp from the log
	now, _ := time.Parse(time.RFC3339, date)
	deadline := now.Add(time.Duration(0-minutes) * time.Minute)

	return func(lines []string, line []byte, _ int) (bool, []string, error) {
		words := strings.Fields(reverse(line))
		lineTime, err := time.Parse(time.RFC3339, words[0])
		if err != nil {
			return true, lines, err
		}
		diff := deadline.Sub(lineTime)
		return diff >= 0, append(lines, string(line)), nil
	}
}

// Check defines function signature to check whether the condition was met
type Check func([]string, []byte, int) (bool, []string, error)

// TailFile trims end of the file by defined func
func TailFile(filename string, fn Check) error {
	if len(filename) == 0 {
		return errors.New("you must provide the path to a file")
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
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
			return err
		}

		// make sure start position can never be less than 0
		if startPos == 0 {
			break
		}

		b := make([]byte, 1)
		_, err = file.ReadAt(b, startPos)
		line = append(line, b...)
		if err != nil {
			return err
		}

		// ignore if first character being read is a newline
		if offset == int64(-1) && string(b) == "\n" {
			offset--
			line = []byte{}
			continue
		}

		// if the character is a newline add to the number of lines read
		if string(b) == "\n" {
			cond, resultLines, err = fn(resultLines, line, lines)
			if err != nil {
				return err
			}

			if cond {
				break
			}
			lines++
			line = []byte{}
		}

		offset--
	}

	return writeLines(reverseLines(resultLines), filename)
}

func writeLines(lines []string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, line := range lines {
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}
	}

	return nil
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
