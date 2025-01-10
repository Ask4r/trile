package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/ask4r/trile/bot"
	"github.com/ask4r/trile/convert"
	"github.com/ask4r/trile/hash"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

const (
	DATA_DIR = "data"
)

var (
	loConv *convert.LOConv
)

func handleError(b *bot.Bot, u *tgbotapi.Update, err error) {
	chatId := u.Message.Chat.ID
	text := fmt.Sprintf("Ooops...\n[Error] %q", err)
	err = b.SendMsg(chatId, text)
	if err != nil {
		log.Printf("Event your err handlers need handlers... err: %q", err)
		return
	}
}

func handleUpdate(b *bot.Bot, u *tgbotapi.Update) {
	m := u.Message
	if m == nil {
		log.Print("No message sent")
		return
	}
	d := m.Document
	if d == nil {
		log.Print("No document sent")
		return
	}
	chatId := m.Chat.ID

	fn := DATA_DIR + "/" + hash.SNow()
	ext := path.Ext(d.FileName)
	if ext == ".pdf" {
		b.SendMsg(chatId, "Cannot convert PDF to PDF")
		return
	}
	destext := ".pdf"
	srcfn := fn + ext
	destfn := fn + destext
	targetfn := strings.TrimSuffix(d.FileName, ext) + destext

	err := b.FetchDoc(d, srcfn)
	if err != nil {
		log.Print("Could not fetch document")
		handleError(b, u, err)
		return
	}
	defer os.Remove(srcfn)

	if loConv == nil {
		log.Panic("No running LO instance. Fatal")
	}
	err = loConv.OfficeToPdf(srcfn, DATA_DIR)
	if err != nil {
		log.Printf("Could not parse document, err: %q", err)
		handleError(b, u, err)
		return
	}
	defer os.Remove(destfn)

	err = b.SendFile(chatId, destfn, targetfn)
	if err != nil {
		log.Printf("Could not send file, err: %q", err)
		handleError(b, u, err)
		return
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panic("Cannot read .env")
	}
	APIKey := os.Getenv("BOT_API_KEY")

	loConv = convert.New()
	if loConv == nil {
		log.Panic("No running LO instance. Fatal")
	}
	defer loConv.Shutdown()
	log.Printf("LO started successfully")

	b := bot.New(APIKey)
	b.Handle(handleUpdate)
}
