package main

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/ask4r/trile/bot"
	"github.com/ask4r/trile/convert"
	"github.com/ask4r/trile/hash"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

const (
	DATA_DIR = "data"
)

func handleUpdate(b *bot.Bot, lo *convert.LOConv, u *tgbotapi.Update) error {
	// Get message data
	m := u.Message
	if m == nil {
		return errors.New("no message received")
	}
	d := m.Document
	if d == nil {
		return errors.New("no document received")
	}
	chatId := m.Chat.ID

	// Get temporary filenames
	fn := DATA_DIR + "/" + hash.NowString()
	ext := path.Ext(d.FileName)
	if ext == ".pdf" {
		err := b.SendMsg(chatId, "Cannot convert PDF to PDF")
		return err
	}
	destext := ".pdf"
	srcfn := fn + ext
	destfn := fn + destext
	docname := strings.TrimSuffix(d.FileName, ext) + destext

	// Fetch document
	err := b.FetchDoc(d, srcfn)
	if err != nil {
		return err
	}
	defer func() {
		err := os.Remove(srcfn)
		if err != nil {
			log.Printf("could not cleanup file: \"%s\"", srcfn)
		}
	}()

	// Convert document
	err = lo.OfficeToPdf(srcfn, DATA_DIR)
	if err != nil {
		return err
	}
	defer func() {
		err := os.Remove(destfn)
		if err != nil {
			log.Printf("could not cleanup file: \"%s\"", destfn)
		}
	}()

	// Send document back
	err = b.SendFile(chatId, destfn, docname)
	if err != nil {
		return errors.Wrap(err, "could not send file")
	}

	return nil
}

func main() {
	// Retrieve Env data
	err := godotenv.Load()
	if err != nil {
		log.Panicf("cannot read .env: %v", err)
	}
	apiKey := os.Getenv("BOT_API_KEY")
	if apiKey == "" {
		log.Panic("cannot retrieve environment variable \"BOT_API_KEY\"")
	}

	// Start LO instance
	log.Print("starting new LibreOffice instance")
	lo, err := convert.New()
	if err != nil {
		log.Panicf("could not start LO: %v", err)
	}
	defer func() {
		err := lo.Shutdown()
		if err != nil {
			log.Panicf("could not shutdown LO: %v", err)
		}
	}()
	log.Print("LO started successfully")

	// Connect to Bot API
	b, err := bot.New(apiKey)
	if err != nil {
		log.Panicf("cound not create bot: %v", err)
	}
	bname := b.API.Self.UserName
	log.Printf("Authorized on account @%s \"https://t.me/%s\"", bname, bname)

	// Handle Bot updates
	chMsgErr := make(chan error)
	go func() {
		for err := range chMsgErr {
			log.Printf("message handler error: %v", err)
		}
	}()
	b.Handle(func(u *tgbotapi.Update) {
		go func() {
			err := handleUpdate(b, lo, u)
			if err != nil {
				chMsgErr <- err
			}
		}()
	})
}
