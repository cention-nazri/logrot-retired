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
	quit    chan struct{}
}

// WriteTo sets the log output to the given file and reopen the file on SIGHUP.
func WriteTo(name string) *LogRot {
	return rotateOn(name, syscall.SIGHUP)
}

// WriteAllTo sets the log output, os.Stdout and os.Stderr to the given file and reopen the file on SIGHUP.
func WriteAllTo(name string) *LogRot {
	lr := WriteTo(name)
	os.Stdout = lr.LogFile
	os.Stderr = lr.LogFile
	return lr
}

// rotateOn rotates the log file on the given signals
func rotateOn(name string, sig os.Signal) *LogRot {
	rl := &LogRot{
		name:    name,
		signal:  sig,
		LogFile: mustOpenFileForAppend(name),
		quit:    make(chan struct{}),
	}
	log.SetOutput(rl.LogFile)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, sig)
	go func() {
		for {
			select {
			case s := <-sigs:
				if s == rl.signal {
					log.Printf("%s received - rotating log file handle on %s\n", s, rl.name)
					rl.rotate()
				}
			case <-rl.quit:
				return
			}
		}
	}()
	return rl
}

func (rl *LogRot) Close() {
	if rl != nil && rl.LogFile != nil {
		rl.quit <- struct{}{}
		rl.LogFile.Close()
	}
}

func (rl *LogRot) rotate() {
	oldLog := rl.LogFile
	rl.LogFile = mustOpenFileForAppend(rl.name)
	log.SetOutput(rl.LogFile)
	oldLog.Close()
}
