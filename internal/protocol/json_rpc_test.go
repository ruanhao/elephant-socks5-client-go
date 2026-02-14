package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentHello(t *testing.T) {
	agentHello := NewAgentHello()
	buf, err := agentHello.ToJson()
	assert.NoError(t, err)
	bufString := string(buf)
	t.Log(bufString)
	assert.Contains(t, bufString, `"jsonrpc":"2.0"`)
	assert.Contains(t, bufString, `"method":"agent-hello"`)
	assert.Contains(t, bufString, `"params":`)
	assert.Contains(t, bufString, `"version":`)
	assert.Contains(t, bufString, `"alias":`)
	assert.Contains(t, bufString, `"flowControl":`)
	assert.Contains(t, bufString, `"reverse":`)
	assert.Contains(t, bufString, `"shell":`)
	assert.Contains(t, bufString, `"myip":`)
	assert.Contains(t, bufString, `"seq":`)
}
