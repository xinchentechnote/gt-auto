package tcp_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xinchentechnote/gt-auto/pkg/config"
	"github.com/xinchentechnote/gt-auto/pkg/proto"
	"github.com/xinchentechnote/gt-auto/pkg/tcp"
)

type dummyMessage struct {
	Content string
}

func (d *dummyMessage) MsgType() uint32 {
	return 999999
}

func (d *dummyMessage) Encode() ([]byte, error) {
	return []byte(d.Content), nil
}

func (d *dummyMessage) Decode(data []byte) error {
	d.Content = string(data)
	return nil
}

func init() {
	proto.RegisterMessage(999999, func() proto.Message {
		return &dummyMessage{}
	})
}

func TestOmsTgwIntegration(t *testing.T) {
	config, err := config.ParseConfig("../config/testdata/gw-auto.toml")
	require.NoError(t, err)
	// Start TGW server in a goroutine
	tgw, err := tcp.CreateSimulator[proto.Message](config.Simulators[1])
	require.NoError(t, err)
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
	oms, err := tcp.CreateSimulator[proto.Message](config.Simulators[0])
	require.NoError(t, err)
	err = oms.Start()
	require.NoError(t, err)
	defer oms.Close()

	//Oms send a test message
	testMsg := &dummyMessage{
		Content: "LOGIN",
	}
	err = oms.Send(testMsg)
	require.NoError(t, err)
	time.Sleep(1000 * time.Millisecond)
	//Tgw receive the message
	resp, err := tgw.Receive()
	assert.Nil(t, err)
	assert.Equal(t, "LOGIN", resp.(*dummyMessage).Content)
}
