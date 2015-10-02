// Package logrot handles log rotation on SIGHUP.
package logrot

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func mustOpenFileForAppend(name string) *os.File {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	return f
}

// LogRot represents log file that will be reopened on a given signal.
type LogRot struct {
	name    string
	LogFile *os.File
	signal  os.Signal
}

// WriteTo sets the log output to the given file and reopen the file on SIGHUP.
func WriteTo(name string) *LogRot {
	return rotateOn(name, syscall.SIGHUP)
}

// rotateOn rotates the log file on the given signals
func rotateOn(name string, sig os.Signal) *LogRot {
	rl := &LogRot{
		name:    name,
		signal:  sig,
		LogFile: mustOpenFileForAppend(name),
	}
	log.SetOutput(rl.LogFile)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, sig)
	go func() {
		for s := range sigs {
			switch s {
			case rl.signal:
				log.Printf("%s received - rotating log file handle on %s\n", s, rl.name)
				oldLog := rl.LogFile
				rl.LogFile = mustOpenFileForAppend(rl.name)
				log.SetOutput(rl.LogFile)
				oldLog.Close()
			}
		}
	}()
	return rl
}

func (rl *LogRot) Close() {
	if rl != nil && rl.LogFile != nil {
		rl.LogFile.Close()
	}
}
