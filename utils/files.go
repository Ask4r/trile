package utils

import (
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
