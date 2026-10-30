
package handlers

import (
	"fmt"
	"sync"
	"io"

	"github.com/gin-gonic/gin"
)

var (
	clients   = make(map[chan string]bool)
	clientsMu sync.Mutex
)

// Broadcast sends an update notification to all connected users.
func Broadcast(event string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		select {
		case client <- event:
		default:
			// Avoid blocking if a client is slow.
		}
	}
}

// Events handles the Server-Sent Events connection.
func Events(c *gin.Context) {
	client := make(chan string, 10)

	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, client)
		close(client)
		clientsMu.Unlock()
	}()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	c.Stream(func(w io.Writer) bool {
		select {
		case event := <-client:
			fmt.Fprintf(w, "data: %s\n\n", event)
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}
