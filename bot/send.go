package bot

import (
	"io"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

func (b *Bot) SendMsg(chatId int64, text string) error {
	msg := tgbotapi.NewMessage(chatId, text)
	_, err := b.API.Send(msg)
	if err != nil {
		return errors.Wrap(err, "could not send message")
	}
	return nil
}

func (b *Bot) SendFile(chatId int64, fn, docname string) error {
	f, err := os.Open(fn)
	if err != nil {
		return errors.Wrap(err, "could not open file")
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		return errors.Wrap(err, "could not read file")
	}
	doc := tgbotapi.FileBytes{
		Name:  docname,
		Bytes: bytes,
	}

	msg := tgbotapi.NewDocument(chatId, doc)
	_, err = b.API.Send(msg)
	if err != nil {
		return errors.Wrap(err, "could not send document")
	}

	return nil
}
