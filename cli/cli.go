package cli

import (
	"flag"
	"fmt"
)

const (
	LOG_LEVEL = "info"
	LOG_FILE  = "/var/log/trile-bot/trile.log"
	TMP_DIR   = "/var/tmp/trile-bot"
)

type CLIParams struct {
	LogLevel *string
	LogFile  *string
	TmpDir   *string
}

func Parse() (CLIParams, error) {
	log_level := flag.String("log-level", "info", "Log level: \"debug\", \"info\", \"warn\", \"error\". Defaults to \"info\"")
	log_file := flag.String("log-file", LOG_FILE, "Logs output file path. \"stdout\" for console output. Default to \""+LOG_FILE+"\"")
	tmp_dir := flag.String("tmp-dir", TMP_DIR, "Dir for necessary temporary files. Defaults to \""+TMP_DIR+"\"")

	flag.Parse()

	params := CLIParams{
		LogLevel: log_level,
		LogFile:  log_file,
		TmpDir:   tmp_dir,
	}

	err := validate(params)
	if err != nil {
		return params, err
	}

	return params, nil
}

func validate(params CLIParams) error {
	log_level := *params.LogLevel
	if log_level != "debug" &&
		log_level != "info" &&
		log_level != "warn" &&
		log_level != "error" {
		return fmt.Errorf("Unknown log level: \"%s\"", log_level)
	}
	return nil
}
