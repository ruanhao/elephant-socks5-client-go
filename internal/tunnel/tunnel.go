package tunnel

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ruanhao/elephant/internal/config"
	"github.com/ruanhao/elephant/internal/protocol"
)

type Tunnel struct {
	helloC chan struct{}

	outboundChannel        chan *bytes.Buffer
	inboundChannel         <-chan *bytes.Buffer
	conn                   *websocket.Conn
	readBuffer             bytes.Buffer
	pendingJsonRpcRequests map[string]string
	mu                     sync.Mutex
	context                context.Context
	cancelFunc             context.CancelCauseFunc
	closeOnce              sync.Once
}

func NewTunnel() *Tunnel {
	t := &Tunnel{}
	t.helloC = make(chan struct{})
	t.outboundChannel = make(chan *bytes.Buffer, 32)
	t.pendingJsonRpcRequests = make(map[string]string)
	t.context, t.cancelFunc = context.WithCancelCause(context.Background())
	return t
}

func (t *Tunnel) createInboundChannel(conn *websocket.Conn) <-chan *bytes.Buffer {
	ch := make(chan *bytes.Buffer)
	t.inboundChannel = ch
	go func() {
		defer close(ch)
		for {
			_, data, err := conn.ReadMessage()

			if err != nil {
				slog.Error("Error during reading websocket message", "error", err)
				return
			}
			select {
			case ch <- bytes.NewBuffer(data):
			case <-t.context.Done():
				slog.Info("Stop reading from websocket, tunnel context done")
				return
			}
		}
	}()
	return ch
}

func (t *Tunnel) setPendingJsonRpcResponseMethod(id string, method string) {
	t.mu.Lock()
	func() {
		defer t.mu.Unlock()
		t.pendingJsonRpcRequests[id] = method
	}()
}

func (t *Tunnel) getPendingJsonRpcResponseMethod(id string) string {
	t.mu.Lock()
	defer t.mu.Unlock()
	defer delete(t.pendingJsonRpcRequests, id)
	slog.Info("check pending requests", "total", len(t.pendingJsonRpcRequests))
	return t.pendingJsonRpcRequests[id]
}

func (t *Tunnel) Start() {
	defer close(t.helloC)
	dialer := websocket.Dialer{
		TLSClientConfig: config.TlsConfig,
	}
	header := http.Header{}
	header.Set("User-Agent", "Elephant/go-client")
	slog.Info("Connecting to elephant server", "url", config.GetWebSocketURL())
	conn, _, err := dialer.Dial(config.GetWebSocketURL(), header)

	if err != nil {
		slog.Error("Error during connecting to elephant", "err", err)
		startLater()
		return
	}

	slog.Info("Tunnel Active")

	t.conn = conn
	t.createInboundChannel(conn)
	t.SpawnHandleInbound()
	t.SpawnHandleOutbound()
	t.SendAgentHello()

	select {
	case <-t.helloC:
		slog.Info("Tunnel hello response received, tunnel setup complete")
	case <-time.After(10 * time.Second):
		slog.Error("Timeout waiting for tunnel hello response, closing tunnel")
		t.close(fmt.Errorf("timeout waiting for tunnel hello response"))
		return
	}

}

func (t *Tunnel) SendAgentHello() {
	hello := protocol.NewAgentHello()
	frame := hello.ToFrame()
	slog.Info("Sending agent hello", "size", frame.Len())
	t.outboundChannel <- frame
	t.setPendingJsonRpcResponseMethod(hello.Id, protocol.JSONRPC_METHOD_SERVER_HELLO)
}

func (t *Tunnel) handleJsonRPC(payload *bytes.Buffer) {
	slog.Info("Handling JSON-RPC message", "payload", payload.String())
	jsonRpc, err := protocol.FromByteBuf(payload)
	if err != nil {
		slog.Error("Error during parsing JSON-RPC message", "error", err)
		return
	}

	method := jsonRpc.Method
	slog.Info("Received JSON-RPC message", "method", method)
	if method == "" {
		method = t.getPendingJsonRpcResponseMethod(jsonRpc.Id)
		if method == "" { // No expected method found for this id
			slog.Error("Received JSON-RPC response with unknown id", "id", jsonRpc.Id)
			return
		}
	}
	switch method {
	case protocol.JSONRPC_METHOD_SERVER_HELLO:
		t.helloC <- struct{}{}
	case protocol.JSONRPC_METHOD_ECHO_REQUEST:
		t.handleEchoRequest(jsonRpc)
	default:
		slog.Warn("Received JSON-RPC message with unhandled method", "method", method)
	}

}

func (t *Tunnel) handleFrame(frame *protocol.Frame) {
	slog.Info("Handling frame", "frame", frame)
	op, payload := frame.Op(), frame.Payload
	switch op {
	case protocol.OP_CONTROL:

		slog.Debug("Received control message", "payload", payload.String())
		t.handleJsonRPC(payload)
	case protocol.OP_DATA:
		slog.Debug("Received data message", "size", payload.Len())
	default:
		slog.Error("Received frame with unknown op", "op", op)
	}

}

func (t *Tunnel) tunnelRead(byteBuf *bytes.Buffer) {
	// Put data into readBuffer first
	t.readBuffer.Write(byteBuf.Bytes())

	// Process complete frames
	for {
		frame := protocol.ExtractFrame(&t.readBuffer)
		if frame == nil {
			// No more complete frames to process
			break
		}
		t.handleFrame(frame)
	}

	if t.readBuffer.Len() == 0 {
		t.readBuffer = bytes.Buffer{} // Reset buffer to free memory
	}
}

func (t *Tunnel) SpawnHandleInbound() {
	go func() {
		for {
			select {
			case byteBuf, ok := <-t.inboundChannel:
				if !ok {
					t.close(errors.New("inbound channel closed"))
					return
				}
				t.tunnelRead(byteBuf)

			case <-t.context.Done():
				slog.Info("Tunnel handle context done, exiting handle INBOUND loop")
				return
			}

		}
	}()
}
func (t *Tunnel) SpawnHandleOutbound() {
	go func() {
		defer close(t.outboundChannel)
		for {
			select {
			case outboundData, ok := <-t.outboundChannel:
				if !ok {
					slog.Info("Outbound byte buffer channel closed, exiting handle loop")
					return
				}
				if outboundData.Len() == 0 {
					slog.Warn("Attempting to send empty outbound data, skipping")
					continue
				}
				slog.Info("Sending outbound data", "size", outboundData.Len())
				err := t.conn.WriteMessage(websocket.BinaryMessage, outboundData.Bytes())
				if err != nil {
					t.close(err)
					return
				}

			case <-t.context.Done():
				slog.Info("Tunnel handle context done, exiting handle OUTBOUND loop")
				return
			}

		}
	}()
}

func (t *Tunnel) close(err error) {
	t.closeOnce.Do(func() {
		if t.conn != nil {
			err := t.conn.Close()
			if err != nil {
				slog.Error("Error during closing connection", "error", err)
			}
		}
		t.cancelFunc(err)
		startLater()
	})
}

func (t *Tunnel) handleEchoRequest(rpc *protocol.JsonRPC) {
	slog.Info("Received echo request", "id", rpc.Id)
	t.outboundChannel <- rpc.Response().ToFrame()
	slog.Info("Sent echo response", "id", rpc.Id)
}

func startLater() {
	go func() {
		slog.Info("Retrying connection in 5 seconds...")
		time.Sleep(5 * time.Second)
		t := NewTunnel()
		t.Start()
	}()
}
