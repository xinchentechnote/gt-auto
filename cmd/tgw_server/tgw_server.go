package main

import (
	"log"

	"github.com/xinchentechnote/gt-auto/pkg/tcp"
)

func main() {
	// Start the TGWServer
	tgwServer := tcp.TgwSimulator{ListenAddress: ":10776"}
	go func() {
		if err := tgwServer.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	select {} // Keep the server running indefinitely
}
