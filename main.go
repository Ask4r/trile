package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ask4r/trile/domains/bot"
	"github.com/ask4r/trile/domains/converter"
	"github.com/ask4r/trile/lib/files"
	"github.com/ask4r/trile/pipelines/convert"
	"github.com/joho/godotenv"
)

func main() {
	// Load config
	err := godotenv.Load()
	if err != nil {
		// Nothing is wrong actually. Variables may be defined in the environment.
	}
	conf, err := LoadConfig()
	if err != nil {
		fmt.Printf("Config load error: \"%v\"", err)
		return
	}

	// Files setup
	err = os.MkdirAll(conf.TmpDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Could not create TMP_DIR \"%s\": %v\n", conf.TmpDir, err)
		return
	}
	err = files.CreateFilePath(conf.LogFile)
	if err != nil {
		fmt.Printf("Could not create LOG_FILE \"%s\": %v\n", conf.LogFile, err)
		return
	}

	// Logger init
	var log_file *os.File
	if conf.LogFile == "stdout" {
		log_file = os.Stdout
	} else {
		log_file, err := os.OpenFile(conf.LogFile, os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			fmt.Printf("Could not acess log file \"%s\": %v\n", conf.LogFile, err)
			return
		}
		defer files.CloseRC(log_file)
	}
	logger := slog.New(slog.NewJSONHandler(log_file, &slog.HandlerOptions{Level: conf.LogLevel}))
	slog.SetDefault(logger)

	// Start LO instance
	lo, err := converter.New(conf.TmpDir)
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
	b, err := bot.New(conf.BotApiKey)
	if err != nil {
		fmt.Printf("Could not create bot: %v\n", err)
	}
	bname := b.API.Self.UserName
	fmt.Printf("Authorized on account @%s \"https://t.me/%s\"\n", bname, bname)

	// Start pipeline. Handle updates
	convert.Start(b, lo, conf.TmpDir)
}
