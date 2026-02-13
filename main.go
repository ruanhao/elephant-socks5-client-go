package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"bitbucket.org/JeremySchlatter/go-atexit"
	"github.com/ruanhao/elephant/internal/config"
	"github.com/ruanhao/elephant/internal/tunnel"
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
	gitCommit   = "unknown"
	enableLog   bool
	serverHost  string
	serverPort  int
	alias       string
	port        int
	global      bool
	debugPort   int
	quiet       bool
	flowControl bool
)

func init() {
	flag.Usage = func() {
		_, err := fmt.Fprintf(os.Stderr, "Elephant SOCKS5 Tunnel Client (commit: %s)\n\nUsage:\n  elephant [flags]\n\nFlags:\n", gitCommit)
		if err != nil {
			log.Fatalf("Failed to write to stderr: %v", err)
		}
		flag.PrintDefaults()
	}
	flag.BoolVarP(&enableLog, "log", "l", false, "enable logging to file")
	// flag.StringVarP(&logFile, "log-file", "f", ".log/elephant.log", "log file path")
	// flag.IntVarP(&logMaxSize, "log-max-size", "s", 10, "max size in MB of the log file before it gets rotated")
	// flag.IntVarP(&logMaxBackups, "log-max-backups", "b", 10, "max number of old log files to keep")
	flag.BoolVarP(&quiet, "quiet", "q", false, "quiet mode")
	flag.StringVarP(&serverHost, "server-host", "s", "20.205.132.65", "Elephant server host")
	flag.IntVarP(&serverPort, "server-port", "", 443, "Elephant server port")
	flag.StringVarP(&alias, "alias", "a", "", "alias name")
	flag.IntVarP(&port, "socks5-listening-port", "p", 1080, "SOCKS5 listening port")
	flag.BoolVarP(&global, "global", "g", false, "SOCKS5 port listening on all interfaces")
	flag.IntVarP(&debugPort, "debug-port", "", 6060, "debug HTTP server port (0 to disable)")
	flag.BoolVarP(&flowControl, "flow-control", "f", false, "enable flow control mode")

}

func setupLog() {
	if quiet {
		log.SetOutput(io.Discard)
		return
	}
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

func setupAppConfig() {
	config.AppConfig.Global = global
	config.AppConfig.ServerHost = serverHost
	config.AppConfig.ServerPort = serverPort
	config.AppConfig.Alias = alias
	config.AppConfig.Socks5ListeningPort = port
	config.AppConfig.DebugHTTPPort = debugPort
	config.AppConfig.FlowControl = flowControl
}

func main() {
	atexit.TrapSignals()
	defer atexit.CallExitFuncs()

	atexit.Run(func() {
		time.Sleep(500 * time.Millisecond) // give time to flush logs
		//_, err := fmt.Fprintf(os.Stderr, "elephant exits\n")
		//if err != nil {
		//	return
		//}
	})

	flag.Parse()
	setupLog()
	setupAppConfig()

	slog.Info("hello, elephant!!")
	slog.Info("App config", "config", config.AppConfig)

	t := tunnel.NewTunnel()
	t.Start()

	// Wait for interrupt signal to gracefully shut down the application
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	sig := <-sigCh
	slog.Info("Received signal, shutting down...", "signal", sig)

}
