package main

import (
	"fmt"
	"log"

	"github.com/xinchentechnote/gt-auto/pkg/tcp"
)

func main() {
	// Start the OMSClient
	omsClient := &tcp.OmsSimulator{ServerAddress: "localhost:10776"}
	if err := omsClient.Start(); err != nil {
		log.Fatalf("Client error: %v", err)
	}

	// Send a "LOGIN" request
	if err := omsClient.Send("LOGIN"); err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	// Receive the response
	response, err := omsClient.Receive()
	if err != nil {
		log.Fatalf("Failed to receive response: %v", err)
	}

	fmt.Println("Received from server:", response)

	// Close the client
	if err := omsClient.Close(); err != nil {
		log.Printf("Failed to close client: %v", err)
	}
}
