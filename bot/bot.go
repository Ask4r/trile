package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

type Bot struct {
	API *tgbotapi.BotAPI
	ch  tgbotapi.UpdatesChannel
}

func New(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, errors.Wrap(err, "could not authorize bot")
	}
	api.Debug = false

	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 60
	ch := api.GetUpdatesChan(cfg)

	return &Bot{API: api, ch: ch}, nil
}

func (b *Bot) Handle(handle func(u *tgbotapi.Update)) {
	for u := range b.ch {
		handle(&u)
	}
}

func (b *Bot) GetMsgCommand(m *tgbotapi.Message) string {
	if m == nil {
		return ""
	}
	var text string
	var ents []tgbotapi.MessageEntity
	if m.Text != "" {
		text = m.Text
		ents = m.Entities
	} else if m.Caption != "" {
		text = m.Caption
		ents = m.CaptionEntities
	} else {
		return ""
	}

	for _, e := range ents {
		if e.Type == "bot_command" {
			return text[e.Offset : e.Offset+e.Length]
		}
	}
	return ""
}
