package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/ask4r/trile/bot"
	"github.com/ask4r/trile/download"
	"github.com/ask4r/trile/hash"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/tgulacsi/agostle/converter"
)

// const PPTX_MIME = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
const PPTX_MIME = "application/vnd.ms-powerpoint;charset=UTF-8"

func getUrl(b *tgbotapi.BotAPI, d *tgbotapi.Document) (string, error) {
	id := d.FileID
	f, err := b.GetFile(tgbotapi.FileConfig{FileID: string(id)})
	if err != nil {
		return "", err
	}
	return f.Link(b.Token), nil
}

func loadFile(url, ext string) (string, error) {
	p := "data/" + hash.SNow() + ext
	return p, download.ToFile(p, url)
}

func pptxToPdf(filename, dest string, ctx context.Context) error {
	// converter.Converter
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	return converter.OfficeToPdf(ctx, dest, r, PPTX_MIME)
	// return converter.OfficeToPdf(ctx, dest, r, "pptx")
}

func stripExt(p string) string {
	ext := path.Ext(p)
	return p[:len(p)-len(ext)]
}

func handleUpdate(b *bot.Bot, u *tgbotapi.Update) {
	m := u.Message
	if m == nil {
		return
	}
	d := m.Document
	if d == nil {
		return
	}

	url, err := getUrl(b.Bot, d)
	if err != nil {
		return
	}
	filename, err := loadFile(url, ".pptx")
	if err != nil {
		return
	}

	name := d.FileName
	dest := "parsed/" + stripExt(name) + ".pdf"
	fmt.Printf("parsing `%s` to `%s`\n", filename, dest)
	err = pptxToPdf(filename, dest, b.Ctx)
	if err != nil {
		return
	}

	fmt.Printf("file `%s` converted to `%s`\n", name, dest)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panic("Cannot read .env")
	}

	APIKey := os.Getenv("BOT_API_KEY")

	b := bot.New(APIKey)

	b.Handle(handleUpdate)
}
