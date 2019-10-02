package compress

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkliczewski/fileops/fileops"
)

// Compress tars provided source directory and uses Check function to filter logs
func Compress(filename string, source string, fn fileops.Check) error {
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

		if filepath.Ext(relPath) == ".log" && fn != nil {
			err = fileops.TailFile(path, fn)
			if err != nil {
				// TODO remove once we understand which files should be parsed
				fmt.Println(path)
			}

			info, err = os.Stat(path)
			if err != nil {
				return err
			}
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
