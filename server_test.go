package jsonrpc

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewServer(t *testing.T) {
	server := New()
	server.Register(&adder{})
	go require.NoErrorf(t, server.Start(3124), "server failed to start up successfully")
}
