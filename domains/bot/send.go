package bot

import (
	"io"
	"os"

	"github.com/ask4r/trile/lib/files"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

func (b *Bot) send(c tgbotapi.Chattable) error {
	_, err := b.API.Send(c)
	if err != nil {
		return errors.Wrap(err, "could not send message")
	}
	return nil
}

func loadFileBytes(fn, docname string) (*tgbotapi.FileBytes, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, errors.Wrap(err, "could not open file")
	}
	defer files.CloseRC(f)
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "could not read file")
	}
	return &tgbotapi.FileBytes{Name: docname, Bytes: bytes}, nil
}

func (b *Bot) SendText(chatId int64, text string) error {
	msg := tgbotapi.NewMessage(chatId, text)
	return b.send(msg)
}

func (b *Bot) SendMarkdown(chatId int64, text string) error {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	return b.send(msg)
}

func (b *Bot) ReplyText(chatId int64, msgId int, text string) error {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ReplyToMessageID = msgId
	return b.send(msg)
}

func (b *Bot) SendFile(chatId int64, fn, docname string) error {
	fbytes, err := loadFileBytes(fn, docname)
	if err != nil {
		return err
	}
	doc := tgbotapi.NewDocument(chatId, fbytes)
	return b.send(doc)
}

func (b *Bot) ReplyFile(chatId int64, msgId int, fn, docname string) error {
	fbytes, err := loadFileBytes(fn, docname)
	if err != nil {
		return err
	}
	doc := tgbotapi.NewDocument(chatId, fbytes)
	doc.ReplyToMessageID = msgId
	return b.send(doc)
}
