package compress

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkliczewski/fileops/fileops"
)

var (
	date = "2019-09-26T12:08:46Z"
)

// Compress tars provided source directory and uses numLines or mins to filter
func Compress(filename string, source string, numLines int, mins int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		err = file.Close()
	}()

	gz := gzip.NewWriter(file)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		info, err = filter(relPath, path, numLines, mins, info)

		// fill in header info using func FileInfoHeader
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// ensure header has relative file path
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// if path is a dir, dont continue
		if info.Mode().IsDir() {
			return nil
		}

		// add file to tar
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		if _, err := io.Copy(tw, srcFile); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func filter(relPath string, path string, numLines int, mins int, info os.FileInfo) (os.FileInfo, error) {
	var fn fileops.Check = nil
	var err error = nil
	if strings.Contains(relPath, "kubelet") || strings.Contains(relPath, "NetworkManager") {
		fn, err = getPartial(numLines, mins, path, true)
		if err != nil {
			return info, err
		}
	} else if filepath.Ext(relPath) == ".log" {
		fn, err = getPartial(numLines, mins, path, false)
		if err != nil {
			return info, err
		}
	}

	if fn != nil {
		err = fileops.TailFile(path, fn)
		if err != nil {
			// TODO remove once we understand which files should be parsed
			fmt.Println(path)
		}
		info, err = os.Stat(path)
		if err != nil {
			return info, err
		}
	}
	return info, nil
}

func getPartial(numLines int, mins int, filename string, journal bool) (fileops.Check, error) {
	var fn fileops.Check = nil
	var err error = nil

	// mock current data base on last time stamp from the log
	now, _ := time.Parse(time.RFC3339, date)
	deadline := now.Add(time.Duration(0-mins) * time.Minute)

	if mins > 0 {
		if journal {
			fn, err = fileops.JournalFunc(filename, deadline)
		} else {
			fn = fileops.DatePartial(deadline)
		}
	} else if numLines > 0 {
		fn = fileops.LimitPartial(numLines)
	}
	return fn, err
}
