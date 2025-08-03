package tcp_test

import (
	"bytes"

	"github.com/xinchentechnote/fin-proto-go/codec"
	risk_bin "github.com/xinchentechnote/fin-proto-go/risk-bin/messages"
)

type dummyMessage struct {
	Content string
}

func (d *dummyMessage) MsgType() uint32 {
	return 999999
}

func (d *dummyMessage) Encode(buf *bytes.Buffer) error {
	err := codec.PutString[uint16](buf, d.Content)
	return err
}

func (d *dummyMessage) Decode(buf *bytes.Buffer) error {
	data, err := codec.GetString[uint16](buf)
	d.Content = data
	return err
}

func init() {
	risk_bin.RegistryRcBinaryMsgTypeFactory(999999, func() codec.BinaryCodec {
		return &dummyMessage{}
	})
}

// func TestOmsTgwIntegration(t *testing.T) {
// 	config, err := config.ParseConfig("../config/testdata/gw-auto.toml")
// 	require.NoError(t, err)
// 	// Start TGW server in a goroutine
// 	tgw, err := tcp.CreateSimulator[codec.BinaryCodec](config.Simulators[0])
// 	require.NoError(t, err)
// 	go func() {
// 		err := tgw.Start()
// 		require.NoError(t, err)
// 	}()
// 	t.Cleanup(func() {
// 		tgw.Close()
// 	})
// 	// Wait for server to start
// 	time.Sleep(100 * time.Millisecond)

// 	// Start OMS client
// 	oms, err := tcp.CreateSimulator[codec.BinaryCodec](config.Simulators[1])
// 	require.NoError(t, err)
// 	err = oms.Start()
// 	require.NoError(t, err)
// 	defer oms.Close()

// 	//Oms send a test message
// 	testMsg := &dummyMessage{
// 		Content: "LOGIN",
// 	}
// 	err = oms.Send(testMsg)
// 	require.NoError(t, err)
// 	time.Sleep(1000 * time.Millisecond)
// 	//Tgw receive the message
// 	resp, err := tgw.Receive()
// 	assert.Nil(t, err)
// 	assert.Equal(t, "LOGIN", resp.(*dummyMessage).Content)
// }
