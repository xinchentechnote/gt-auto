package tcp

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
	fin_codec "github.com/xinchentechnote/fin-proto-go/codec"
	"github.com/xinchentechnote/gt-auto/pkg/codec"
)

// Simulator interface defines the methods for both OMS and TGW simulators
type Simulator[T fin_codec.BinaryCodec] interface {
	Start() error
	Send(interface{}, fin_codec.BinaryCodec) error
	//SendFromJSON to send JSON-like map,it should implement convert JSON-like map to T
	SendFromJSON(message map[string]interface{}) error
	Receive() (T, error)
	GetCodec() codec.MessageCodec
	Close() error
}

// OmsSimulator simulates the OMS client
type OmsSimulator[T fin_codec.BinaryCodec] struct {
	ServerAddress string
	conn          net.Conn
	queue         *goconcurrentqueue.FIFO
	Codec         codec.MessageCodec
	Framer        codec.Framer
}

// TgwSimulator simulates the TGW server
type TgwSimulator[T fin_codec.BinaryCodec] struct {
	ListenAddress string
	listener      net.Listener
	stopChan      chan struct{}
	queue         *goconcurrentqueue.FIFO
	Codec         codec.MessageCodec
	Framer        codec.Framer
	conn          net.Conn
}

func (sim *OmsSimulator[T]) GetCodec() codec.MessageCodec {
	return sim.Codec
}

// Start connects to the TGWServer
func (sim *OmsSimulator[T]) Start() error {
	sim.queue = goconcurrentqueue.NewFIFO()
	var err error
	sim.conn, err = net.DialTimeout("tcp", sim.ServerAddress, 5*time.Second)
	if err != nil {
		log.Printf("failed to connect to server: %s", err)
		return fmt.Errorf("failed to connect to server: %w", err)

	}
	log.Printf("Connected to TGW server at %s", sim.ServerAddress)
	go func() {
		if err := sim.receive0(); err != nil {
			log.Printf("receive0 error: %v", err)
		}
	}()
	return nil
}

// Send sends a message to the server
func (sim *OmsSimulator[T]) Send(ext interface{}, message fin_codec.BinaryCodec) error {
	data, e := sim.Codec.Encode(ext, message)
	if e != nil {
		return fmt.Errorf("failed to encode message: %w", e)
	}
	return sim.sendByte(data)
}

func (sim *OmsSimulator[T]) sendByte(message []byte) error {
	_, err := sim.conn.Write(message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (sim *OmsSimulator[T]) SendFromJSON(message map[string]interface{}) error {
	data, e := sim.Codec.EncodeJSONMap(message)
	if e != nil {
		return fmt.Errorf("failed to encode message: %w", e)
	}
	return sim.sendByte(data)
}

// Receive waits for a response from the server
func (sim *OmsSimulator[T]) Receive() (T, error) {
	msg, err := sim.queue.Dequeue()
	if err != nil {
		var zero T
		return zero, fmt.Errorf("error dequeuing message: %w", err)
	}
	return msg.(T), nil
}

// Receive waits for a response from the server
func (sim *OmsSimulator[T]) receive0() error {
	data, err := sim.Framer.ReadFrame(sim.conn)
	if err != nil {
		return fmt.Errorf("failed to receive message: %w", err)
	}
	_, msg, e := sim.Codec.Decode(data)
	if e != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}
	e1 := sim.queue.Enqueue(msg)
	if e1 != nil {
		return fmt.Errorf("failed to enqueue message: %w", e1)
	}

	return nil
}

// Close closes the OMSClient connection
func (sim *OmsSimulator[T]) Close() error {
	return sim.conn.Close()
}

// GetCodec returns the message codec used by the simulator
func (sim *TgwSimulator[T]) GetCodec() codec.MessageCodec {
	return sim.Codec
}

// Start listens for incoming connections on the TGWServer
func (sim *TgwSimulator[T]) Start() error {
	var err error
	sim.listener, err = net.Listen("tcp", sim.ListenAddress)
	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}
	log.Printf("TGW server started on %s", sim.ListenAddress)
	sim.stopChan = make(chan struct{})
	sim.queue = goconcurrentqueue.NewFIFO()
	go func() {
		<-sim.stopChan
		sim.listener.Close()
	}()

	for {
		conn, err := sim.listener.Accept()
		sim.conn = conn
		if err != nil {
			select {
			case <-sim.stopChan:
				log.Println("TGW server shutting down.")
				return nil
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}
		go sim.handleClient(conn)
	}
}

// Handle incoming client connections and put messages in the queue
func (sim *TgwSimulator[T]) handleClient(conn net.Conn) {
	defer conn.Close()

	for {
		data, err := sim.Framer.ReadFrame(conn)
		if err != nil {
			log.Printf("Error decoding message: %v", err)
			continue
		}
		_, msg, e := sim.Codec.Decode(data)
		if e != nil {
			log.Printf("Error decoding message: %v", e)
			continue
		}
		e1 := sim.queue.Enqueue(msg)
		if e1 != nil {
			log.Printf("Error enqueuing message: %v", e1)
			continue
		}
	}
}

// Send sends a message to the client
func (sim *TgwSimulator[T]) Send(ext interface{}, message fin_codec.BinaryCodec) error {
	data, e := sim.Codec.Encode(ext, message)
	if e != nil {
		return fmt.Errorf("failed to encode message: %w", e)
	}
	return sim.sendByte(data)
}

func (sim *TgwSimulator[T]) sendByte(message []byte) error {
	_, err := sim.conn.Write(message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (sim *TgwSimulator[T]) SendFromJSON(message map[string]interface{}) error {
	bytes, e := sim.Codec.EncodeJSONMap(message)
	if e != nil {
		return fmt.Errorf("failed to encode message: %w", e)
	}
	return sim.sendByte(bytes)
}

// Receive reads the next message from the queue
func (sim *TgwSimulator[T]) Receive() (T, error) {
	msg, err := sim.queue.Dequeue()
	if err != nil {
		var zero T
		return zero, fmt.Errorf("error dequeuing message: %w", err)
	}
	return msg.(T), nil
}

// Close shuts down the TGWServer
func (sim *TgwSimulator[T]) Close() error {
	close(sim.stopChan)
	return nil
}
