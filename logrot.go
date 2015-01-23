// Package logrot handles log rotation on SIGUSR1.
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
	logFile *os.File
	signal  os.Signal
}

// WriteTo sets the log output to the given file and reopen the file on SIGUSR1.
func WriteTo(name string) *LogRot {
	return rotateOn(name, syscall.SIGUSR1)
}

// rotateOn rotates the log file on the given signals
func rotateOn(name string, sig os.Signal) *LogRot {
	rl := &LogRot{
		name:    name,
		signal:  sig,
		logFile: mustOpenFileForAppend(name),
	}
	log.Println("Logging to", name)
	log.SetOutput(rl.logFile)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, sig)
	go func() {
		for s := range sigs {
			switch s {
			case rl.signal:
				log.Printf("%s received - rotating log file handle on %s\n", s, rl.name)
				oldLog := rl.logFile
				rl.logFile = mustOpenFileForAppend(rl.name)
				log.SetOutput(rl.logFile)
				oldLog.Close()
			}
		}
	}()
	return rl
}

func (rl *LogRot) Close() {
	if rl != nil && rl.logFile != nil {
		rl.logFile.Close()
	}
}
