package main

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/ask4r/trile/bot"
	"github.com/ask4r/trile/convert"
	"github.com/ask4r/trile/hash"
	"github.com/ask4r/trile/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

const (
	DATA_DIR = "data"
	LOG_FILE = ".local/state/trile/logs/trile.log"
)

func handleUpdate(b *bot.Bot, lo *convert.LOConv, u *tgbotapi.Update) {
	// fmt.Printf("Incoming message: %+v\n", u.Message)
	// Get message data
	m := u.Message
	if m == nil {
		slog.Info("no message received", "update", u)
		return
	}
	chatId := m.Chat.ID
	slog.Debug("new message", "message", m)
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
		err := b.ReplyText(chatId, m.MessageID, "Cannot convert PDF to PDF")
		if err != nil {
			slog.Error("could not send message", "error", err, "chatId", chatId)
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
	defer utils.RemoveFile(srcfn)

	// Convert document
	err = lo.OfficeToPdf(srcfn, DATA_DIR)
	if err != nil {
		slog.Error("could not convert file", "error", err, "file", srcfn)
		return
	}
	defer utils.RemoveFile(destfn)

	// Send document back
	err = b.ReplyFile(chatId, m.MessageID, destfn, docname)
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
		fmt.Printf("Cannot read .env: %v\n", err)
		return
	}
	apiKey := os.Getenv("BOT_API_KEY")
	if apiKey == "" {
		fmt.Print("Cannot retrieve environment variable \"BOT_API_KEY\"\n")
		return
	}

	// Init logger
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Cannot use log file: cannot retrieve HOME dir: %v\n", err)
		return
	}
	logfn := path.Join(homedir, LOG_FILE)
	logf, err := os.OpenFile(logfn, os.O_RDWR|os.O_APPEND, 0o666)
	if err != nil {
		fmt.Printf("Could not acess log file: %v\n", err)
		return
	}
	defer utils.CloseRC(logf)
	fmt.Printf("Logs will be stored in \"%s\"\n", logfn)
	logger := slog.New(slog.NewJSONHandler(logf,
		&slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// Start LO instance
	lo, err := convert.New()
	if err != nil {
		fmt.Printf("Could not start LO: %v\n", err)
		return
	}
	defer func() {
		if err := lo.Shutdown(); err != nil {
			fmt.Printf("Could not shutdown LO: %v\n", err)
			return
		}
	}()

	// Connect to Bot API
	b, err := bot.New(apiKey)
	if err != nil {
		fmt.Printf("Could not create bot: %v\n", err)
	}
	bname := b.API.Self.UserName
	fmt.Printf("Authorized on account @%s \"https://t.me/%s\"\n", bname, bname)

	// Handle Bot updates
	b.Handle(func(u *tgbotapi.Update) {
		go handleUpdate(b, lo, u)
	})
}
