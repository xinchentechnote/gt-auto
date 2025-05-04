package tcp

import (
	"fmt"
	"io"
	"log"
	"net"
)

// Simulator interface defines the methods for both OMS and TGW simulators
type Simulator interface {
	Start() error
	Send(message string) error
	Receive() (string, error)
	Close() error
}

// OmsSimulator simulates the OMS client
type OmsSimulator struct {
	ServerAddress string
	conn          net.Conn
}

// TgwSimulator simulates the TGW server
type TgwSimulator struct {
	ListenAddress string
	listener      net.Listener
	stopChan      chan struct{}
}

// Start connects to the TGWServer
func (c *OmsSimulator) Start() error {
	var err error
	c.conn, err = net.Dial("tcp", c.ServerAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	log.Printf("Connected to TGW server at %s", c.ServerAddress)
	return nil
}

// Send sends a message to the server
func (c *OmsSimulator) Send(message string) error {
	_, err := c.conn.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// Receive waits for a response from the server
func (c *OmsSimulator) Receive() (string, error) {
	buf := make([]byte, 1024)
	n, err := c.conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to receive message: %w", err)
	}
	return string(buf[:n]), nil
}

// Close closes the OMSClient connection
func (c *OmsSimulator) Close() error {
	return c.conn.Close()
}

// Start listens for incoming connections on the TGWServer
func (s *TgwSimulator) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.ListenAddress)
	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}
	log.Printf("TGW server started on %s", s.ListenAddress)
	s.stopChan = make(chan struct{})

	go func() {
		<-s.stopChan
		s.listener.Close()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopChan:
				log.Println("TGW server shutting down.")
				return nil
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}
		go s.handleClient(conn)
	}
}

// Handle incoming client connections and put messages in the queue
func (s *TgwSimulator) handleClient(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from connection: %v", err)
			}
			break
		}
		msg := string(buf[:n])
		fmt.Println("Received message:", msg)

		resp := s.ProcessMessage(msg)

		_, err1 := conn.Write([]byte(resp))
		if err1 != nil {
			log.Printf("Error writing to connection: %v", err)
			break
		}
	}
}

// Receive reads the next message from the queue
func (s *TgwSimulator) Receive() (string, error) {
	return "", nil
}

// ProcessMessage processes the received message (for example, a simple echo)
func (s *TgwSimulator) ProcessMessage(message string) string {
	// Here we can add more complex processing, but we'll just echo the message for now
	return "Processed: " + message
}

// Close shuts down the TGWServer
func (s *TgwSimulator) Close() error {
	close(s.stopChan)
	return nil
}
