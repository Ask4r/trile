package utils

import (
	"errors"
	"io"
	"log/slog"
	"os"
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
	return errors.Is(err, os.ErrNotExist)
}
