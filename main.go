package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logFile       = ".log/elephant.log"
	logMaxSize    = 10 // MB
	logMaxBackups = 10
	logMaxAge     = 30 // days
)

var (
	enableLog bool
)

func init() {
	flag.Usage = func() {
		_, err := fmt.Fprintf(os.Stderr, "Elephant SOCKS5 Tunnel Client\n\nUsage:\n  elephant [flags]\n\nFlags:\n")
		if err != nil {
			log.Fatalf("Failed to write to stderr: %v", err)
		}
		flag.PrintDefaults()
	}
	flag.BoolVarP(&enableLog, "log", "l", false, "enable logging to file")
	// flag.StringVarP(&logFile, "log-file", "f", ".log/elephant.log", "log file path")
	// flag.IntVarP(&logMaxSize, "log-max-size", "s", 10, "max size in MB of the log file before it gets rotated")
	// flag.IntVarP(&logMaxBackups, "log-max-backups", "b", 10, "max number of old log files to keep")

}

func setupLog() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	if enableLog {
		// make sure the log directory exists
		if dir := filepath.Dir(logFile); dir != "" {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				log.Fatalf("Failed to create log directory %s: %v", dir, err)
			}
		}
		lj := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    logMaxSize,
			MaxBackups: logMaxBackups,
			MaxAge:     logMaxAge, // days
			Compress:   true,
		}
		log.SetOutput(io.MultiWriter(os.Stdout, lj))
	} else {
		log.SetOutput(os.Stdout)
	}
}

func main() {
	flag.Parse()
	setupLog()
	slog.Info("hello, elephant!!")
}
