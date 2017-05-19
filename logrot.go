// Package logrot handles log rotation on SIGHUP.
package logrot

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Logger interface {
	SetOutput(io.Writer)
}

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
	quit    chan struct{}

	loggers []Logger

	captureStdout bool
	captureStderr bool
}

// WriteTo sets the log output to the given file and reopen the file on SIGHUP.
func WriteTo(name string, loggers ...Logger) *LogRot {
	return rotateOn(name, syscall.SIGHUP, loggers...)
}

// WriteAllTo sets the log output, os.Stdout and os.Stderr to the given file and reopen the file on SIGHUP.
func WriteAllTo(name string, loggers ...Logger) *LogRot {
	lr := WriteTo(name, loggers...)
	lr.CaptureStdout()
	lr.CaptureStderr()
	return lr
}

// rotateOn rotates the log file on the given signals
func rotateOn(name string, sig os.Signal, loggers ...Logger) *LogRot {
	rl := &LogRot{
		name:    name,
		signal:  sig,
		logFile: mustOpenFileForAppend(name),
		quit:    make(chan struct{}),
		loggers: loggers,
	}

	rl.setOutput()

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

func (rl *LogRot) setOutput() {
	log.SetOutput(rl.logFile)
	for _, l := range rl.loggers {
		l.SetOutput(rl.logFile)
	}
	if rl.captureStdout {
		os.Stdout = rl.logFile
	}
	if rl.captureStderr {
		os.Stderr = rl.logFile
	}
}

func (rl *LogRot) Close() {
	if rl != nil && rl.logFile != nil {
		rl.quit <- struct{}{}
		rl.logFile.Close()
	}
}

func (rl *LogRot) rotate() {
	oldLog := rl.logFile
	rl.logFile = mustOpenFileForAppend(rl.name)
	rl.setOutput()
	oldLog.Close()
}

func (rl *LogRot) CaptureStdout() {
	rl.captureStdout = true
	os.Stdout = rl.logFile
}

func (rl *LogRot) CaptureStderr() {
	rl.captureStderr = true
	os.Stderr = rl.logFile
}
