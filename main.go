package main

import (
	"log"
	"log/slog"
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
	LOG_FILE = ".local/state/trile/logs/trile.log"
)

func cleanupFile(fn string) {
	err := os.Remove(fn)
	if err != nil {
		slog.Error("could not cleanup file", "error", err, "file", fn)
	}
}

func handleUpdate(b *bot.Bot, lo *convert.LOConv, u *tgbotapi.Update) {
	// Get message data
	m := u.Message
	if m == nil {
		slog.Info("no message received", "update", u)
		return
	}
	chatId := m.Chat.ID
	d := m.Document
	if d == nil {
		slog.Info("no document received", "chatId", chatId)
		return
	}

	// Temporary filenames
	basefn := DATA_DIR + "/" + hash.NowString()
	ext := path.Ext(d.FileName)
	destext := ".pdf"
	if ext == destext {
		slog.Info("tried to convert wrong file extension", "fromExt", ext, "toExt", destext)
		err := b.SendMsg(chatId, "Cannot convert PDF to PDF")
		if err != nil {
			slog.Error("could not send message", "error", err, "chatId", chatId)
			return
		}
		return
	}
	srcfn := basefn + ext
	destfn := basefn + destext // expected out filename
	docname := strings.TrimSuffix(d.FileName, ext) + destext

	// Fetch document
	err := b.FetchDoc(d, srcfn)
	if err != nil {
		slog.Error("could not fetch document", "error", err, "document", d, "chatId", chatId)
		return
	}
	defer cleanupFile(srcfn)

	// Convert document
	err = lo.OfficeToPdf(srcfn, DATA_DIR)
	if err != nil {
		slog.Error("could not convert file", "error", err, "file", srcfn)
		return
	}
	defer cleanupFile(destfn)

	// Send document back
	err = b.SendFile(chatId, destfn, docname)
	if err != nil {
		slog.Error("could not send file", "error", err, "file", destfn, "chatId", chatId)
		return
	}

	slog.Info("successfully converted document", "document", d, "chatId", chatId)
	return
}

func main() {
	var err error

	// Retrieve Env data
	err = godotenv.Load()
	if err != nil {
		log.Panicf("cannot read .env: %v", err)
	}
	apiKey := os.Getenv("BOT_API_KEY")
	if apiKey == "" {
		log.Panic("cannot retrieve environment variable \"BOT_API_KEY\"")
	}

	// Init logger
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("cannot use log file: cannot retrieve HOME dir: %v", err)
	}
	logfn := path.Join(homedir, LOG_FILE)
	logf, err := os.OpenFile(logfn, os.O_RDWR, 0o666)
	if err != nil {
		log.Panicf("could not acess log file: %v", err)
	}
	defer func() {
		err := logf.Close()
		if err != nil {
			log.Printf("could not close log file: %v", err)
		}
	}()
	log.Printf("logs will be stored in \"%s\"", logfn)
	logger := slog.New(slog.NewJSONHandler(logf,
		&slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Start LO instance
	slog.Info("starting new LibreOffice instance")
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
	log.Printf("authorized on account @%s \"https://t.me/%s\"", bname, bname)

	// Handle Bot updates
	b.Handle(func(u *tgbotapi.Update) {
		go handleUpdate(b, lo, u)
	})
}
