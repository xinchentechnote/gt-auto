package tcp_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinchentechnote/gt-auto/pkg/tcp"
)

func TestOmsTgwIntegration(t *testing.T) {
	// Start TGW server in a goroutine
	tgw := &tcp.TgwSimulator{ListenAddress: "localhost:9001"}
	go func() {
		err := tgw.Start()
		require.NoError(t, err)
	}()
	t.Cleanup(func() {
		tgw.Close()
	})
	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Start OMS client
	oms := &tcp.OmsSimulator{ServerAddress: "localhost:9001"}
	err := oms.Start()
	require.NoError(t, err)
	defer oms.Close()

	// Send a test message
	testMsg := "LOGIN"
	err = oms.Send(testMsg)
	require.NoError(t, err)

	// Receive (not implemented to echo yet, so no response expected)
	resp, err := oms.Receive()
	assert.Nil(t, err)                        // expect read to fail or be empty
	assert.Equal(t, "Processed: LOGIN", resp) // or "" as response
}
