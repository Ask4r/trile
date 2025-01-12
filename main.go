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

func ReplyText(b *bot.Bot, chatId int64, msgId int, text string) {
	err := b.ReplyText(chatId, msgId, text)
	if err != nil {
		slog.Error("could not send message", "error", err, "chatId", chatId)
	}

}

func sendStartMsg(b *bot.Bot, chatId int64) {
	startMessage := "Hi, I'm *Trile*\\!\n\nI can convert some document types\\! " +
		"Try sending me `.pptx`, `.docx`, `.xlsx` or any other document and I will convert it to `.pdf`\\. " +
		"You can also specify other target file extensions with commands like /pdf and I'll try to convert them too\\!\n\n" +
		"Or you can reply to other's messages with documents and choose target file types with commands\\!"
	err := b.SendMarkdown(chatId, startMessage)
	if err != nil {
		slog.Error("could not send message", "error", err, "chatId", chatId)
	}
}

func handleUpdate(b *bot.Bot, lo *convert.LOConv, u *tgbotapi.Update) {
	// Get message data
	m := u.Message
	if m == nil {
		slog.Info("no message received", "update", u)
		return
	}
	slog.Debug("new message", "message", m)
	chatId := m.Chat.ID
	cmd := b.GetMsgCommand(m)
	if cmd == "/start" {
		sendStartMsg(b, chatId)
	}
	d := m.Document
	if d == nil {
		repMsg := m.ReplyToMessage
		if repMsg != nil && repMsg.Document != nil {
			d = repMsg.Document
		} else {
			slog.Info("no document received", "chatId", chatId)
			return
		}
	}

	// Temporary filenames
	basefn := DATA_DIR + "/" + hash.NowString()
	var target string
	if cmd == "" {
		target = "pdf"
	} else {
		target = strings.TrimPrefix(cmd, "/")
	}
	srcext := path.Ext(d.FileName)
	destext := "." + target
	if srcext == destext {
		slog.Info("tried to convert to same file extension", "fromExt", srcext, "toExt", destext)
		reply := fmt.Sprintf("That's %s to %s...", srcext, destext)
		ReplyText(b, chatId, m.MessageID, reply)
		return
	}
	srcfn := basefn + srcext
	destfn := basefn + destext // expected out filename
	docname := strings.TrimSuffix(d.FileName, srcext) + destext

	// Fetch document
	err := b.FetchDoc(d, srcfn)
	if err != nil {
		slog.Error("could not fetch document", "error", err, "document", d, "chatId", chatId)
		reply := fmt.Sprintf("Could not fetch %s. Sorry!", d.FileName)
		ReplyText(b, chatId, m.MessageID, reply)
		return
	}
	defer utils.RemoveFile(srcfn)

	// Convert document
	err = lo.OfficeToExt(srcfn, DATA_DIR, target)
	if err != nil {
		slog.Error("could not convert file", "error", err, "file", srcfn)
		reply := "Something definetly went wrong. I did my best. It doesn't work. Trust me."
		ReplyText(b, chatId, m.MessageID, reply)
		return
	}
	if utils.PathExist(destfn) {
		slog.Error("impossible conversion", "fromExt", srcext, "toExt", destext, "file", srcfn)
		reply := fmt.Sprintf("Cannot convert %s to %s. That's witchery!", srcext, destext)
		ReplyText(b, chatId, m.MessageID, reply)
		return

	}
	defer utils.RemoveFile(destfn)

	// Send document back
	err = b.ReplyFile(chatId, m.MessageID, destfn, docname)
	if err != nil {
		slog.Error("could not send file", "error", err, "file", destfn, "chatId", chatId)
		reply := "No, seriously... I converted the doc, it was ready, everything was good, but it didn't sent! What!?"
		ReplyText(b, chatId, m.MessageID, reply)
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
	logf, err := os.OpenFile(logfn, os.O_WRONLY|os.O_APPEND, 0o666)
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
