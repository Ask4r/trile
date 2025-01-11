package bot

import (
	"io"
	"os"

	"github.com/ask4r/trile/fetch"
	"github.com/ask4r/trile/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

func (b *Bot) getDocUrl(d *tgbotapi.Document) (string, error) {
	id := d.FileID
	f, err := b.API.GetFile(tgbotapi.FileConfig{FileID: string(id)})
	if err != nil {
		return "", errors.Wrap(err, "could not retrieve doc url")
	}
	return f.Link(b.API.Token), nil
}

func (b *Bot) FetchDoc(d *tgbotapi.Document, dest string) error {
	url, err := b.getDocUrl(d)
	if err != nil {
		return err
	}
	err = fetch.ToFile(dest, url)
	if err != nil {
		return errors.Wrap(err, "could not fetch document")
	}
	return nil
}

func (b *Bot) GetDocStream(d *tgbotapi.Document) (io.ReadCloser, error) {
	url, err := b.getDocUrl(d)
	if err != nil {
		return nil, err
	}
	r, err := fetch.ToStream(url)
	if err != nil {
		return nil, errors.Wrap(err, "could not obtain doc stream")
	}
	return r, nil
}

func loadFileBytes(fn, docname string) (*tgbotapi.FileBytes, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, errors.Wrap(err, "could not open file")
	}
	defer utils.CloseRC(f)
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "could not read file")
	}
	return &tgbotapi.FileBytes{Name: docname, Bytes: bytes}, nil
}
