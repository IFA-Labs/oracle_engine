package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"oracle_engine/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PriceStreamer handles SSE connections for price updates
type PriceStreamer struct {
	// Main price channel that receives price updates
	priceCh chan models.Issuance

	// Client management
	clients    map[string]chan models.Issuance
	clientDone map[string]chan struct{}
	mu         sync.RWMutex

	// Logger
	logger *zap.Logger
}

// NewPriceStreamer creates a new price streaming service
func NewPriceStreamer(priceCh chan models.Issuance, logger *zap.Logger) *PriceStreamer {
	return &PriceStreamer{
		priceCh:    priceCh,
		clients:    make(map[string]chan models.Issuance),
		clientDone: make(map[string]chan struct{}),
		logger:     logger,
	}
}

// Start begins the main forwarding process from the price channel to clients
func (ps *PriceStreamer) Start() {
	go ps.distributePrice()
}

// Stop gracefully shuts down the price streamer
func (ps *PriceStreamer) Stop() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Signal all clients to disconnect
	for clientID := range ps.clientDone {
		close(ps.clientDone[clientID])
	}

	// Close all client channels
	for clientID := range ps.clients {
		close(ps.clients[clientID])
	}

	// Clear maps
	ps.clients = make(map[string]chan models.Issuance)
	ps.clientDone = make(map[string]chan struct{})
}

// @Summary Model Stream price updates
// @Description Server-Sent Events stream of price updates, have a retry mechanism in place for break
// @Tags prices
// @Produce text/event-stream
// @Success 200 {string} models.Issuance "SSE stream"
// @Router /prices/stream [get]
func (ps *PriceStreamer) HandleStream(c *gin.Context) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Flush headers immediately
	if f, ok := c.Writer.(http.Flusher); ok {
		f.Flush()
	}

	// Generate client ID
	clientID := uuid.New().String()

	// Create a channel for this client
	clientChan := make(chan models.Issuance, 10) // Buffer for 10 price updates

	// Register client
	ps.registerClient(clientID, clientChan)

	// Make sure to unregister when done
	defer ps.unregisterClient(clientID)

	// Important: detect when client disconnects
	clientGone := c.Request.Context().Done()

	// Set up streaming
	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			ps.logger.Debug("Client disconnected", zap.String("clientID", clientID))
			return false

		case price, ok := <-clientChan:
			if !ok {
				ps.logger.Debug("Client channel closed", zap.String("clientID", clientID))
				return false
			}

			// Format and send the message
			err := ps.writeSSEMessage(w, "price", price)
			if err != nil {
				ps.logger.Error("Failed to write SSE message",
					zap.Error(err),
					zap.String("clientID", clientID))
				return false
			}

			// Flush data immediately
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

			return true
		}
	})
}

// distributePrice forwards prices from the main channel to all client channels
func (ps *PriceStreamer) distributePrice() {
	for price := range ps.priceCh {
		// Make a copy of the current clients under read lock
		ps.mu.RLock()
		clientChannels := make([]chan models.Issuance, 0, len(ps.clients))
		for _, ch := range ps.clients {
			clientChannels = append(clientChannels, ch)
		}
		ps.mu.RUnlock()

		// Send the price to each client without holding the lock
		for _, clientChan := range clientChannels {
			select {
			case clientChan <- price:
				// Successfully sent
			default:
				// Channel full, skip this update for this client
				ps.logger.Debug("Skipped price update for client (buffer full)")
			}
		}
	}
}

// registerClient adds a new client to receive price updates
func (ps *PriceStreamer) registerClient(clientID string, ch chan models.Issuance) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Create done channel for signaling
	done := make(chan struct{})

	// Store references
	ps.clients[clientID] = ch
	ps.clientDone[clientID] = done

	ps.logger.Debug("Client registered", zap.String("clientID", clientID))
}

// unregisterClient removes a client and cleans up resources
func (ps *PriceStreamer) unregisterClient(clientID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Signal the client goroutine to stop
	if done, exists := ps.clientDone[clientID]; exists {
		close(done)
		delete(ps.clientDone, clientID)
	}

	// Close and remove the client channel
	if ch, exists := ps.clients[clientID]; exists {
		close(ch)
		delete(ps.clients, clientID)
		ps.logger.Debug("Client unregistered", zap.String("clientID", clientID))
	}
}

// writeSSEMessage formats and writes an SSE message
func (ps *PriceStreamer) writeSSEMessage(w io.Writer, event string, data interface{}) error {
	// Marshal the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write event and data
	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData)
	return err
}

// GetClientCount returns the number of active clients
func (ps *PriceStreamer) GetClientCount() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return len(ps.clients)
}
