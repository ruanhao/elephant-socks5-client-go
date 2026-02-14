package config

import (
	"crypto/tls"
	"fmt"
	"sync/atomic"
)

var (
	Version = "unknown"
	counter atomic.Int64
)

type Config struct {
	ServerHost          string
	ServerPort          int
	Global              bool
	Alias               string
	Socks5ListeningPort int
	DebugHTTPPort       int
	FlowControl         bool
}

var AppConfig = Config{
	ServerHost:          "",
	ServerPort:          0,
	Alias:               "",
	Global:              false,
	Socks5ListeningPort: 0,
	DebugHTTPPort:       0,
	FlowControl:         false,
}

var TlsConfig = &tls.Config{
	InsecureSkipVerify: true,
}

func GetWebSocketURL() string {
	return fmt.Sprintf("wss://%s:%d/elephant/ws", AppConfig.ServerHost, AppConfig.ServerPort)
}

func GetCounter() int64 {
	defer counter.Add(1)
	return counter.Load()
}
