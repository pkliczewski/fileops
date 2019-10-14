package compress

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
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

		if info.Size() != 0 {
			info, err = filter(relPath, path, numLines, mins, info)
		}

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

		// copy file contents
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
	var fn Check = nil
	var err error = nil
	// filter journal based logs
	if strings.Contains(relPath, "kubelet") || strings.Contains(relPath, "NetworkManager") {
		fn, err = getPartial(numLines, mins, path, true)
		if err != nil {
			return info, err
		}
		// filter all pod logs but not qemu, cdi-apiserver, cdi-deployment and cdi-uploadproxy which are problematic to parse
	} else if filepath.Ext(relPath) == ".log" && !strings.Contains(relPath, "qemu") &&
		!strings.Contains(relPath, "cdi-apiserver") && !strings.Contains(relPath, "cdi-deployment") &&
		!strings.Contains(relPath, "cdi-uploadproxy") {
		fn, err = getPartial(numLines, mins, path, false)
		if err != nil {
			return info, err
		}
	}

	if fn != nil {
		// run filtering logic
		err = TailFile(path, fn)
		if err != nil {
			return info, err
		}
		info, err = os.Stat(path)
		if err != nil {
			return info, err
		}
	}
	return info, nil
}

func getPartial(numLines int, mins int, filename string, journal bool) (Check, error) {
	var fn Check = nil
	var err error = nil

	// filter by date
	if mins > 0 {
		deadline := time.Now().Add(time.Duration(0-mins) * time.Minute)
		if journal {
			fn, err = JournalFunc(filename, deadline)
		} else {
			fn = DatePartial(deadline)
		}
		// filter by number of lines
	} else if numLines > 0 {
		fn = LimitPartial(numLines)
	}
	return fn, err
}
