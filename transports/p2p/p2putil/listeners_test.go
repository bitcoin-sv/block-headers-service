package p2putil

import (
	"testing"

	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/rs/zerolog"
)

func TestInitListeners(t *testing.T) {
	// given
	log := zerolog.Nop()

	// when
	listeners, err := InitListeners(&log)

	// then
	assert.NoError(t, err)
	assert.Equal(t, len(listeners), 1)
	assert.Equal(t, listeners[0].Addr().Network(), "tcp")
	assert.Equal(t, listeners[0].Addr().String(), "[::]:8333")
}
