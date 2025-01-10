package bot

import (
	"io"

	"github.com/ask4r/trile/fetch"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) FetchDoc(d *tgbotapi.Document, dest string) error {
	url, err := b.getDocUrl(d)
	if err != nil {
		return err
	}
	return fetch.ToFile(dest, url)
}

func (b *Bot) GetDocStream(d *tgbotapi.Document) (io.ReadCloser, error) {
	url, err := b.getDocUrl(d)
	if err != nil {
		return nil, err
	}
	return fetch.ToStream(url)
}

func (b *Bot) getDocUrl(d *tgbotapi.Document) (string, error) {
	id := d.FileID
	f, err := b.API.GetFile(tgbotapi.FileConfig{FileID: string(id)})
	if err != nil {
		return "", err
	}
	return f.Link(b.API.Token), nil
}
