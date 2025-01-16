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
	LOG_LEVEL = slog.LevelInfo
	DATA_DIR  = "data"
	LOG_FILE  = ".local/state/trile/logs/trile.log"
)

type ConvertUpdate struct {
	FN      string
	DestExt string
	Doc     *tgbotapi.Document
	Msg     *tgbotapi.Message
}

type RespondUpdate struct {
	FN      string
	DocName string
	Doc     *tgbotapi.Document
	Msg     *tgbotapi.Message
}

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

func getMessageDoc(m *tgbotapi.Message) *tgbotapi.Document {
	if m == nil {
		return nil
	} else if m.Document != nil {
		return m.Document
	} else if m.ReplyToMessage == nil {
		return nil
	} else if m.ReplyToMessage.Document != nil {
		return m.ReplyToMessage.Document
	}
	return nil
}

func handleUpdate(b *bot.Bot, convCh chan ConvertUpdate, u *tgbotapi.Update) {
	// Get message data
	m := u.Message
	if m == nil {
		slog.Warn("no message received", "update", u)
		return
	}
	slog.Debug("new message", "message", m)
	chatId := m.Chat.ID

	cmd := b.GetMsgCommand(m)
	if cmd == "" {
		return
	} else if cmd == "/start" {
		sendStartMsg(b, chatId)
		return
	}

	d := getMessageDoc(m)
	if d == nil {
		slog.Info("no document received", "chatId", chatId)
		reply := "Where's the doc? WHERE IS MY DOC!?"
		ReplyText(b, chatId, m.MessageID, reply)
		return
	}

	// Paths and extensions
	srcext := path.Ext(d.FileName)
	destext := "." + strings.TrimPrefix(cmd, "/")
	if srcext == destext {
		slog.Info("tried to convert to same file extension", "fromExt", srcext, "toExt", destext)
		reply := fmt.Sprintf("That's %s to %s...", srcext, destext)
		ReplyText(b, chatId, m.MessageID, reply)
		return
	}
	fn := DATA_DIR + "/" + hash.NowString() + srcext

	// Fetch document
	err := b.FetchDoc(d, fn)
	if err != nil {
		slog.Error("could not fetch document", "error", err, "document", d, "chatId", chatId)
		reply := fmt.Sprintf("Could not fetch %s. Sorry!", d.FileName)
		ReplyText(b, chatId, m.MessageID, reply)
		return
	}

	// Pass to conversion
	convCh <- ConvertUpdate{FN: fn, DestExt: destext, Doc: d, Msg: m}
}

// This handler is a BOTTLENECK
func handleConvert(b *bot.Bot, lo *convert.LOConv, respCh chan RespondUpdate, u *ConvertUpdate) {
	chatId := u.Msg.Chat.ID
	msgId := u.Msg.MessageID
	srcext := path.Ext(u.FN)
	srcfn := u.FN
	destfn := strings.TrimSuffix(u.FN, srcext) + u.DestExt

	defer utils.RemoveFile(srcfn)

	// Convert document
	loTarget := strings.TrimPrefix(u.DestExt, ".")
	err := lo.OfficeToExt(srcfn, DATA_DIR, loTarget)
	if err != nil {
		slog.Error("could not convert file", "error", err, "file", srcfn)
		reply := "Something definetly went wrong. I did my best. It doesn't work. Trust me."
		// Don't wait for reply to start a new conversion
		go ReplyText(b, chatId, msgId, reply)
		return
	}
	if !utils.PathExist(destfn) {
		slog.Error("impossible conversion", "fromExt", srcext, "toExt", u.DestExt, "file", srcfn)
		reply := fmt.Sprintf("Cannot convert %s to %s. That's witchery!", srcext, u.DestExt)
		// Don't wait for reply to start a new conversion
		go ReplyText(b, chatId, msgId, reply)
		return
	}

	// Pass to responding
	destDocName := strings.TrimSuffix(u.Doc.FileName, srcext) + u.DestExt
	respCh <- RespondUpdate{FN: destfn, DocName: destDocName, Doc: u.Doc, Msg: u.Msg}
}

func handleRespond(b *bot.Bot, u *RespondUpdate) {
	chatId := u.Msg.Chat.ID
	msgId := u.Msg.MessageID

	defer utils.RemoveFile(u.FN)

	// Send document back
	err := b.ReplyFile(chatId, msgId, u.FN, u.DocName)
	if err != nil {
		slog.Error("could not send file", "error", err, "file", u.FN, "chatId", chatId)
		reply := "No, seriously... I converted the doc, it was ready, everything was good, but it didn't sent! What!?"
		ReplyText(b, chatId, msgId, reply)
		return
	}

	slog.Info("successfully converted document", "document", u.Doc, "chatId", chatId)
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
		&slog.HandlerOptions{Level: LOG_LEVEL}))
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
		}
	}()

	// Connect to Bot API
	b, err := bot.New(apiKey)
	if err != nil {
		fmt.Printf("Could not create bot: %v\n", err)
	}
	bname := b.API.Self.UserName
	fmt.Printf("Authorized on account @%s \"https://t.me/%s\"\n", bname, bname)

	// Handle updates
	// CORE PIPELINE
	// handleUpdate -> handleConvert -> handleRespond
	convCh := make(chan ConvertUpdate)
	respCh := make(chan RespondUpdate)
	go func() {
		// When new file is ready to be responded with
		for u := range respCh {
			go handleRespond(b, &u)
		}
	}()
	go func() {
		// When new file is ready for conversion
		for u := range convCh {
			// LO can only work syncronously
			handleConvert(b, lo, respCh, &u)
		}
	}()
	b.Handle(func(u *tgbotapi.Update) {
		go handleUpdate(b, convCh, u)
	})
}
