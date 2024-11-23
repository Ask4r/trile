package bot

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	Bot *tgbotapi.BotAPI
	ch  tgbotapi.UpdatesChannel
	Ctx context.Context
}

func New(token string) *Bot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account @%s \"https://t.me/%s\"", bot.Self.UserName, bot.Self.UserName)

	bot.Debug = false

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 60

	ch := bot.GetUpdatesChan(cfg)

	return &Bot{Bot: bot, ch: ch, Ctx: ctx}
}

func (b *Bot) Handle(handle func(b *Bot, u *tgbotapi.Update)) {
	for u := range b.ch {
		handle(b, &u)
	}
}
