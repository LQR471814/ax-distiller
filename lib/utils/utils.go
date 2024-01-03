package utils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type indexedPair[T any] struct {
	index int
	value T
}

func ParallelMap[I, O any](inputs []I, mapper func(input I) O) []O {
	collector := make(chan indexedPair[O])
	for i, input := range inputs {
		go func(idx int, input I) {
			collector <- indexedPair[O]{
				index: idx,
				value: mapper(input),
			}
		}(i, input)
	}
	results := make([]O, len(inputs))
	for i := 0; i < len(inputs); i++ {
		pair := <-collector
		results[pair.index] = pair.value
	}
	return results
}

func mapInvisible(r rune) rune {
	if unicode.IsGraphic(r) {
		return r
	}
	return -1
}

// Remove invisible characters from a string.
func RemoveInvisible(text string) string {
	return strings.Map(mapInvisible, text)
}

// taken from https://gist.github.com/paulerickson/6d8650947ee4e3f3dbcc28fde10eaae7
func Unzip(archive *zip.Reader, destination string) error {
	for _, file := range archive.File {
		reader, err := file.Open()
		if err != nil {
			return err
		}
		defer reader.Close()
		path := filepath.Join(destination, file.Name)

		// Remove file if it already exists; no problem if it doesn't; other cases can error out below
		_ = os.Remove(path)
		// Create a directory at path, including parents
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}

		// If file is _supposed_ to be a directory, we're done
		if file.FileInfo().IsDir() {
			continue
		}

		// otherwise, remove that directory (_not_ including parents)
		err = os.Remove(path)
		if err != nil {
			return err
		}

		// and create the actual file.  This ensures that the parent directories exist!
		// An archive may have a single file with a nested path, rather than a file for each parent dir
		writer, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer writer.Close()
		_, err = io.Copy(writer, reader)
		if err != nil {
			return err
		}
	}
	return nil
}
