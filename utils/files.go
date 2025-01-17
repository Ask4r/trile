package utils

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func CloseRC(s io.ReadCloser) {
	err := s.Close()
	if err != nil {
		slog.Error("could not close file", "error", err)
	}
}

func RemoveFile(fn string) {
	err := os.Remove(fn)
	if err != nil {
		slog.Error("could not remove file", "error", err, "file", fn)
	}
}

func PathExist(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func CreateFilePath(path string) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return errors.Wrap(err, "unexpected error: cannot close newly created file")
	}
	return nil
}
