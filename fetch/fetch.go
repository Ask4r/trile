package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

func ToStream(url string) (io.ReadCloser, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "could not GET url")
	}
	if r.StatusCode != http.StatusOK {
		r.Body.Close()
		return nil, fmt.Errorf("bad status: %s", r.Status)
	}
	return r.Body, nil
}

func ToFile(fn, url string) error {
	f, err := os.Create(fn)
	if err != nil {
		return errors.Wrap(err, "could not create file")
	}
	defer f.Close()

	r, err := ToStream(url)
	if err != nil {
		return err
	}
	defer r.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return errors.Wrap(err, "could not write stream to file")
	}

	return nil
}
