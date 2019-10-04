package fileops

import (
	"bufio"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	errParsing = errors.New("Unable to parse start and end date")
)

// LimitPartial allows to provide line limit and return condition function
func LimitPartial(numLines int) Check {
	return func(line string, lines int) (bool, error) {
		return lines >= numLines-1, nil
	}
}

// DatePartial allows to provide date and return condition function
func DatePartial(deadline time.Time) Check {

	return func(line string, _ int) (bool, error) {
		words := strings.Fields(line)
		lineTime, err := time.Parse(time.RFC3339, words[0])
		if err != nil {
			return true, err
		}
		diff := deadline.Sub(lineTime)
		return diff >= 0, nil
	}
}

// JournalDatePartial allows to provide journal end date and return condition function
func JournalDatePartial(end time.Time, deadline time.Time) Check {
	re := regexp.MustCompile("\\w{3}\\s\\w+\\s\\d{2}:\\d{2}:\\d{2}")
	format := "2006 Jan 02 03:04:05 MST"

	return func(line string, _ int) (bool, error) {
		d := re.FindAllString(line, -1)
		if len(d) != 1 {
			return false, errParsing
		}

		// use header end time to provide missing values from timestamp
		zone, _ := end.Zone()
		date := strconv.Itoa(end.Year()) + " " + d[0] + " " + zone
		lineTime, err := time.Parse(format, date)
		if err != nil {
			return true, err
		}
		diff := deadline.Sub(lineTime)
		return diff >= 0, nil
	}
}

// Check defines function signature to check whether the condition was met
type Check func(string, int) (bool, error)

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
			l := reverse(line)
			cond, err = fn(l, lines)
			if err != nil {
				return err
			}

			if cond {
				break
			}
			resultLines = append(resultLines, l)
			lines++
			line = []byte{}
		}

		offset--
	}

	return writeLines(reverseLines(resultLines), filename)
}

// JournalFunc creates function for journal logs
func JournalFunc(finename string, deadline time.Time) (Check, error) {
	file, err := os.Open(finename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	// header start and end dates defined in second line
	reader.ReadLine()
	dateLine, _, err := reader.ReadLine()
	if err != nil {
		return nil, err
	}
	// parse header dates
	re := regexp.MustCompile("\\w{3}\\s\\d{4}-\\d{2}-\\d{2}\\s\\d{2}:\\d{2}:\\d{2}\\s\\w{3}")
	dates := re.FindAllString(string(dateLine), -1)
	if len(dates) != 2 {
		return nil, errParsing
	}

	format := "Thu 2006-01-02 03:04:05 MST"
	end, err := time.Parse(format, dates[1])
	if err != nil {
		return nil, err
	}

	if end.Before(deadline) {
		return nil, nil
	}

	return JournalDatePartial(end, deadline), nil
}

// override file with provided lines
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
