package tunnel

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ruanhao/elephant/internal/config"
)

type Tunnel struct {
	HelloDone bool

	outboundDataChannel chan OutboundData
	conn                *websocket.Conn
}

type WebsocketInboundMessage struct {
	Type int
	Data []byte
	Err  error
}

type OutboundData struct {
	Data []byte
}

func NewTunnel() *Tunnel {
	return &Tunnel{
		HelloDone: false,
	}
}

func handleConnect(conn *websocket.Conn) {
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			slog.Error("Error during closing connection", "error", err)
		}
	}(conn)
	slog.Info("Connected to elephant server")

	// Handle the connection (e.g., read/write messages)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			slog.Error("Error during reading message", "error", err)
			break
		}
		slog.Info("Received message from server", "message", string(message))
		// Process the message as needed
	}
}

func (t *Tunnel) readPump(conn *websocket.Conn) <-chan WebsocketInboundMessage {
	ch := make(chan WebsocketInboundMessage)
	go func() {
		defer close(ch)
		for {
			msgType, data, err := conn.ReadMessage()
			ch <- WebsocketInboundMessage{Type: msgType, Data: data, Err: err}
			if err != nil {
				return
			}
		}
	}()
	return ch
}

func (t *Tunnel) Start() {
	dialer := websocket.Dialer{
		TLSClientConfig: config.TlsConfig,
	}
	header := http.Header{}
	header.Set("User-Agent", "Elephant/go-client")
	slog.Info("Connecting to elephant server", "url", config.GetWebSocketURL())
	conn, _, err := dialer.Dial(config.GetWebSocketURL(), header)

	if err != nil {
		slog.Error("Error during connecting to elephant", "err", err)
		t.SpawnRestart()
		return
	}

	slog.Info("Tunnel Active")

	inboundWebsocketMessageChannel := t.readPump(conn)
	t.outboundDataChannel = make(chan OutboundData, 32)
	t.conn = conn
	t.SpawnHandle(inboundWebsocketMessageChannel)
}

func (t *Tunnel) SpawnHandle(inboundWebsocketMessageChannel <-chan WebsocketInboundMessage) {
	go func() {
		for {
			select {
			case msg, ok := <-inboundWebsocketMessageChannel:
				if !ok {
					slog.Info("Tunnel inactive", "err", msg.Err)
					t.CloseConnection()
					t.SpawnRestart()
					return
				}
				slog.Info("Received inbound websocket message", "type", msg.Type, "data", string(msg.Data))
				// Handle the inbound message as needed

			case outboundData := <-t.outboundDataChannel:
				slog.Info("Sending outbound data", "data", string(outboundData.Data))
				err := t.conn.WriteMessage(websocket.BinaryMessage, outboundData.Data)
				if err != nil {
					slog.Error("Error during writing message", "error", err)
					t.CloseConnection()
				}
			}

		}
	}()
}

func (t *Tunnel) CloseConnection() {
	if t.conn != nil {
		err := t.conn.Close()
		if err != nil {
			slog.Error("Error during closing connection", "error", err)
		}
	}
}

func (t *Tunnel) SpawnRestart() {
	go func() {
		// Retry connection after a delay
		// You can implement exponential backoff or other strategies here
		slog.Info("Retrying connection in 5 seconds...")
		time.Sleep(5 * time.Second)
		t.Start()
	}()
}
