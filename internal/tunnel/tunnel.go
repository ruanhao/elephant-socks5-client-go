package tunnel

import (
	"bytes"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ruanhao/elephant/internal/config"
	"github.com/ruanhao/elephant/internal/protocol"
)

type Tunnel struct {
	HelloDone bool

	outboundDataChannel chan bytes.Buffer
	conn                *websocket.Conn
}

type WebsocketInboundMessage struct {
	Type int
	Data []byte
	Err  error
}

func NewTunnel() *Tunnel {
	return &Tunnel{
		HelloDone: false,
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
	t.outboundDataChannel = make(chan bytes.Buffer, 32)
	t.conn = conn
	t.SpawnHandle(inboundWebsocketMessageChannel)
	t.SendAgentHello()

}

func (t *Tunnel) SendAgentHello() {
	hello := protocol.NewAgentHello()
	frame := hello.ToFrame()
	slog.Info("Sending agent hello", "size", frame.Len())
	t.outboundDataChannel <- frame
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
				slog.Info("Sending outbound data", "size", outboundData.Len())
				err := t.conn.WriteMessage(websocket.BinaryMessage, outboundData.Bytes())
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
