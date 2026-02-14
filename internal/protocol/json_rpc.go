package protocol

import (
	"bytes"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/ruanhao/elephant/internal/config"
	"github.com/ruanhao/elephant/internal/utils"
)

const (
	JSONRPC_METHOD_AGNET_HELLO = "agent-hello"

	JSONRPC_METHOD_SESSION_REQUEST  = "session-request"
	JSONRPC_METHOD_SESSION_RESPONSE = "session-response"

	JSONRPC_METHOD_TERMINATION_REQUEST  = "termination-request"
	JSONRPC_METHOD_TERMINATION_RESPONSE = "termination-response"

	JSONRPC_METHOD_ECHO_REQUEST  = "echo-request"
	JSONRPC_METHOD_ECHO_RESPONSE = "echo-response"
)

type JsonRPC struct {
	JsonRpc string         `json:"jsonrpc"`
	Id      string         `json:"id"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
	Result  map[string]any `json:"result"`
	Error   map[string]any `json:"error"`
}

func (j *JsonRPC) ToJson() ([]byte, error) {
	return json.Marshal(j)
}

func (j *JsonRPC) ToFrame() bytes.Buffer {
	jsonBytes, _ := j.ToJson()
	jsonByteBuf := bytes.NewBuffer(jsonBytes)
	header := NewElephantControlHeader(uint16(jsonByteBuf.Len()))
	headerBuf := header.ToBuffer()

	var frame bytes.Buffer
	frame.Write(headerBuf.Bytes())
	frame.Write(jsonByteBuf.Bytes())
	return frame
}

func NewAgentHello() *JsonRPC {
	return &JsonRPC{
		JsonRpc: "2.0",
		Id:      uuid.New().String(),
		Method:  JSONRPC_METHOD_AGNET_HELLO,
		Params: map[string]any{
			"reverse":     true,
			"shell":       true,
			"myip":        utils.GetOutboundIP(),
			"seq":         config.GetCounter(),
			"alias":       config.AppConfig.Alias,
			"version":     config.Version,
			"flowControl": config.AppConfig.FlowControl,
		},
	}
}
