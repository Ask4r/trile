package convert

import (
	"errors"
	"log"
	"os/exec"
	"strconv"
)

const (
	loExec = "/usr/bin/soffice"
	// loAcceptStr       = "socket,host=127.0.0.1,port=2003;urp;"
	loIsolatedEnvFile = "/tmp/LibreOffice_Conversion_TrileBot"
	loAcceptStr       = "socket,host=127.0.0.1,port=2003,tcpNoDelay=1;urp;StarOffice.ComponentContext"
)

type LOConv struct {
	pid int // -1 if no instance is running yet
}

// Starts background LibreOffice (LO) instance.
// Run in before using any other LO jobs, so they
// are sent to single background instance to avoid
// startup/shutdown penalty for speed (primary) and
// safety (secondary) purposes.
// Return Pid of started instance
func New() *LOConv {
	log.Printf("Starting LibreOffice background instance...")
	cmd := exec.Command(
		loExec,
		"--nodefault",
		"--headless",
		"--norestore",
		"--nocrashreport",
		"--accept="+loAcceptStr,
		"-env:UserInstallation=file://"+loIsolatedEnvFile,
	)
	err := cmd.Start()
	if err != nil {
		log.Printf("Could not start LibreOffice")
		return nil
	}
	pid := cmd.Process.Pid
	return &LOConv{pid: pid}
}

func (lo *LOConv) Shutdown() error {
	if lo.pid == -1 {
		return errors.New("No running instance exists")
	}

	log.Printf("Shutting down LO...")
	cmd := exec.Command("kill", "-9", strconv.Itoa(lo.pid))
	err := cmd.Run()
	if err != nil {
		log.Printf("Could not shutdown LO instance")
		return err
	}

	lo.pid = -1

	return nil
}

// Converts LO-supported file to PDF
func (lo *LOConv) OfficeToPdf(srcfn, outdir string) error {
	if lo.pid == -1 {
		return errors.New("No running instance exists")
	}

	cmd := exec.Command(
		loExec,
		"--convert-to", "pdf", srcfn,
		"--outdir", outdir,
		"-env:UserInstallation=file://"+loIsolatedEnvFile,
	)
	// err := cmd.Start()
	err := cmd.Run()
	if err != nil {
		log.Printf("Could not run LO job")
		return err
	}
	// go cmd.Wait()

	return nil
}
