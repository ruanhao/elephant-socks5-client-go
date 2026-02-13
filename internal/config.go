package internal

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
