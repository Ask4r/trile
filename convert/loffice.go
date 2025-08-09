package convert

import (
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
)

const (
	// loAcceptStr    = "socket,host=127.0.0.1,port=2003;urp;"
	loExec            = "/usr/bin/soffice"
	loIsolatedEnvFile = "/tmp/LibreOffice_Conversion_TrileBot"
	loAcceptStr       = "socket,host=127.0.0.1,port=2003,tcpNoDelay=1;urp;StarOffice.ComponentContext"
)

type LOConv struct {
	pid     int // -1 if no instance is running yet
	tmp_dir string
}

func (lo *LOConv) checkInstance() error {
	if lo.pid == -1 {
		return errors.New("no running LibreOffice instance")
	}
	return nil
}

// Starts background LibreOffice (LO) instance.
// Run it before using any other LO jobs, so they
// are sent to single background instance to avoid
// startup/shutdown penalty for speed (primary) and
// safety (secondary) purposes
func New(tmp_dir string) (*LOConv, error) {
	cmd := exec.Command(loExec,
		"--nodefault",
		"--headless",
		"--norestore",
		"--nocrashreport",
		"--accept="+loAcceptStr,
		"-env:UserInstallation=file://"+loIsolatedEnvFile)
	err := cmd.Start()
	if err != nil {
		return nil, errors.Wrap(err, "cound not run LibreOffice instance")
	}

	pid := cmd.Process.Pid
	return &LOConv{pid, tmp_dir}, nil
}

func (lo *LOConv) Shutdown() error {
	err := lo.checkInstance()
	if err != nil {
		return errors.Wrap(err, "cannot shutdown LO")
	}
	cmd := exec.Command("kill", "-9", strconv.Itoa(lo.pid))
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "cound not shutdown LO")
	}
	lo.pid = -1
	return nil
}

func (lo *LOConv) OfficeToExt(fn, ext string) error {
	err := lo.checkInstance()
	if err != nil {
		return errors.Wrap(err, "cannot convert file")
	}

	cmd := exec.Command(loExec,
		"--convert-to", mapExt(ext), fn,
		"--outdir", lo.tmp_dir,
		"-env:UserInstallation=file://"+loIsolatedEnvFile)
	// err = cmd.Start()
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "could not run conversion")
	}
	// cmd.Wait()

	return nil
}

func mapExt(ext string) string {
	switch ext {
	case "txt":
		return "txt:text"
	default:
		return ext
	}
}
