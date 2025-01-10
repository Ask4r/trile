package bot

import (
	"io"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API *tgbotapi.BotAPI
	ch  tgbotapi.UpdatesChannel
}

func New(token string) *Bot {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bname := api.Self.UserName
	log.Printf("Authorized on account @%s \"https://t.me/%s\"", bname, bname)

	api.Debug = false

	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 60

	ch := api.GetUpdatesChan(cfg)

	return &Bot{API: api, ch: ch}
}

func (b *Bot) Handle(handle func(b *Bot, u *tgbotapi.Update)) {
	for u := range b.ch {
		handle(b, &u)
	}
}

func (b *Bot) SendMsg(chatId int64, text string) error {
	msg := tgbotapi.NewMessage(chatId, text)
	_, err := b.API.Send(msg)
	return err
}

func (b *Bot) SendFile(chatId int64, docfn, docname string) error {
	docr, err := os.Open(docfn)
	if err != nil {
		return err
	}
	defer docr.Close()

	docbytes, err := io.ReadAll(docr)
	if err != nil {
		log.Printf("Could not read file \"%s\"", docfn)
		return err
	}
	doc := tgbotapi.FileBytes{
		Name:  docname,
		Bytes: docbytes,
	}

	msg := tgbotapi.NewDocument(chatId, doc)
	_, err = b.API.Send(msg)

	return err
}
