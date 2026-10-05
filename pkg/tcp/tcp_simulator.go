package tcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
	fin_codec "github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
	"github.com/xinchentechnote/gt-auto/pkg/codec"
)

// Simulator interface defines the methods for both OMS and TGW simulators
type Simulator[T fin_codec.BinaryCodec] interface {
	Start() error
	Send(interface{}, fin_codec.BinaryCodec) error
	//SendFromJSON to send JSON-like map,it should implement convert JSON-like map to T
	SendFromJSON(message map[string]interface{}) error
	// Receive waits up to timeout for the next message from the queue.
	Receive(timeout time.Duration) (T, error)
	GetCodec() codec.MessageCodec
	Close() error
}

// OmsSimulator simulates the OMS client
type OmsSimulator[T fin_codec.BinaryCodec] struct {
	ServerAddress string
	connMu        sync.Mutex
	conn          net.Conn
	queue         *goconcurrentqueue.FIFO
	Codec         codec.MessageCodec
	Framer        codec.Framer
}

// TgwSimulator simulates the TGW server
type TgwSimulator[T fin_codec.BinaryCodec] struct {
	ListenAddress string
	listener      net.Listener
	stopMu        sync.Mutex
	stopChan      chan struct{}
	connMu        sync.Mutex
	// conn is the most recently accepted client connection, used by Send.
	conn   net.Conn
	queue  *goconcurrentqueue.FIFO
	Codec  codec.MessageCodec
	Framer codec.Framer
}

func (sim *OmsSimulator[T]) GetCodec() codec.MessageCodec {
	return sim.Codec
}

// Start connects to the TGWServer
func (sim *OmsSimulator[T]) Start() error {
	sim.queue = goconcurrentqueue.NewFIFO()
	conn, err := net.DialTimeout("tcp", sim.ServerAddress, 5*time.Second)
	if err != nil {
		log.Printf("failed to connect to server: %s", err)
		return fmt.Errorf("failed to connect to server: %w", err)

	}
	sim.connMu.Lock()
	sim.conn = conn
	sim.connMu.Unlock()
	log.Printf("Connected to TGW server at %s", sim.ServerAddress)
	go sim.receiveLoop()
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
	sim.connMu.Lock()
	defer sim.connMu.Unlock()
	if sim.conn == nil {
		return fmt.Errorf("no connection established")
	}
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

// Receive waits up to timeout for the next message from the server.
func (sim *OmsSimulator[T]) Receive(timeout time.Duration) (T, error) {
	return dequeueWithContext[T](sim.queue, timeout)
}

// receiveLoop reads frames from the connection until it is closed or the peer
// disconnects. Per-message decode failures are logged and skipped, as the
// framer already consumed the frame and the stream stays aligned.
func (sim *OmsSimulator[T]) receiveLoop() {
	for {
		err := sim.receive0()
		if err == nil {
			continue
		}
		if errors.Is(err, io.EOF) ||
			errors.Is(err, io.ErrUnexpectedEOF) ||
			errors.Is(err, net.ErrClosed) {
			log.Printf("receive loop stopped: %v", err)
			return
		}
		log.Printf("receive0 error: %v", err)
	}
}

// Receive waits for a response from the server
func (sim *OmsSimulator[T]) receive0() error {
	data, err := sim.Framer.ReadFrame(sim.conn)
	if err != nil {
		return fmt.Errorf("failed to receive message: %w", err)
	}
	_, msg, e := sim.Codec.Decode(data)
	if e != nil {
		return fmt.Errorf("failed to decode message: %w", e)
	}
	log.Printf("Received message: %+v", msg)
	e1 := sim.queue.Enqueue(msg)
	if e1 != nil {
		return fmt.Errorf("failed to enqueue message: %w", e1)
	}

	return nil
}

// Close closes the OMSClient connection. Closing the connection also stops
// the receive loop. It is safe to call before Start.
func (sim *OmsSimulator[T]) Close() error {
	sim.connMu.Lock()
	conn := sim.conn
	sim.connMu.Unlock()
	if conn == nil {
		return nil
	}
	return conn.Close()
}

// GetCodec returns the message codec used by the simulator
func (sim *TgwSimulator[T]) GetCodec() codec.MessageCodec {
	return sim.Codec
}

// Start listens for incoming connections on the TGWServer
func (sim *TgwSimulator[T]) Start() error {
	listener, err := net.Listen("tcp", sim.ListenAddress)
	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}
	sim.stopMu.Lock()
	sim.listener = listener
	sim.stopChan = make(chan struct{})
	sim.stopMu.Unlock()
	log.Printf("TGW server started on %s", sim.ListenAddress)
	sim.queue = goconcurrentqueue.NewFIFO()
	go func() {
		<-sim.stopChan
		listener.Close()
	}()

	for {
		conn, err := sim.listener.Accept()
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
		sim.connMu.Lock()
		sim.conn = conn
		sim.connMu.Unlock()
		go sim.handleClient(conn)
	}
}

// Handle incoming client connections and put messages in the queue
func (sim *TgwSimulator[T]) handleClient(conn net.Conn) {
	defer conn.Close()

	for {
		data, err := sim.Framer.ReadFrame(conn)
		if err != nil {
			// A framing error means the peer disconnected or the stream is
			// desynchronized; stop reading instead of busy-looping.
			log.Printf("client %s disconnected: %v", conn.RemoteAddr(), err)
			return
		}
		_, msg, e := sim.Codec.Decode(data)
		if e != nil {
			log.Printf("Error decoding message: %v", e)
			continue
		}
		log.Printf("Received message: %+v", msg)
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
	sim.connMu.Lock()
	defer sim.connMu.Unlock()
	if sim.conn == nil {
		return fmt.Errorf("no client connection established")
	}
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

// Receive waits up to timeout for the next message from the queue.
func (sim *TgwSimulator[T]) Receive(timeout time.Duration) (T, error) {
	return dequeueWithContext[T](sim.queue, timeout)
}

// dequeueWithContext dequeues the next message, waiting at most timeout
// before giving up so callers cannot block forever on a silent peer.
func dequeueWithContext[T fin_codec.BinaryCodec](queue *goconcurrentqueue.FIFO, timeout time.Duration) (T, error) {
	var zero T
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	item, err := queue.DequeueOrWaitForNextElementContext(ctx)
	if err != nil {
		return zero, fmt.Errorf("no message received within %s: %w", timeout, err)
	}
	msg, ok := item.(T)
	if !ok {
		return zero, fmt.Errorf("unexpected message type %T in queue", item)
	}
	return msg, nil
}

// Close shuts down the TGWServer and drops the accepted client connection.
// It is safe to call before Start or twice.
func (sim *TgwSimulator[T]) Close() error {
	sim.stopMu.Lock()
	if sim.stopChan != nil {
		select {
		case <-sim.stopChan:
			// already closed
		default:
			close(sim.stopChan)
		}
	}
	sim.stopMu.Unlock()

	// Closing the client connection unblocks handleClient and any in-flight
	// reads so goroutines can exit.
	sim.connMu.Lock()
	conn := sim.conn
	sim.connMu.Unlock()
	if conn != nil {
		conn.Close()
	}
	return nil
}
