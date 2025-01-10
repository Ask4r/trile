package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func ToStream(url string) (io.ReadCloser, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		r.Body.Close()
		return nil, fmt.Errorf("bad status: %s", r.Status)
	}
	return r.Body, nil
}

func ToFile(fn, url string) error {
	fh, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer fh.Close()

	r, err := ToStream(url)
	defer r.Close()

	_, err = io.Copy(fh, r)
	if err != nil {
		return err
	}

	return nil
}
