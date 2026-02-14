package tunnel

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClosedChannel(t *testing.T) {
	ch := make(chan bytes.Buffer, 32)
	close(ch)
	v := <-ch
	assert.Equal(t, 0, v.Len())
	t.Logf("v: %v", v)

}
