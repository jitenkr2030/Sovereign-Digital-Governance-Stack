package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/csic/surveillance/internal/domain"
	"github.com/csic/surveillance/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

// WebSocketHandler handles WebSocket connections for real-time data
type WebSocketHandler struct {
	ingestionSvc *service.IngestionService
	alertSubs    map[string]map[chan *domain.Alert]bool
	alertSubMu   sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(ingestionSvc *service.IngestionService) *WebSocketHandler {
	ctx, cancel := context.WithCancel(context.Background())

	return &WebSocketHandler{
		ingestionSvc: ingestionSvc,
		alertSubs:    make(map[string]map[chan *domain.Alert]bool),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// HandleIngestion handles WebSocket connections for market data ingestion
func (h *WebSocketHandler) HandleIngestion(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Track connection
	h.ingestionSvc.IncrementConnectionCount()
	defer h.ingestionSvc.DecrementConnectionCount()

	log.Printf("New ingestion WebSocket connection established. Total: %d", h.ingestionSvc.GetConnectionCount())

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Handle incoming messages
	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			// Read message
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}

			// Parse market data packet
			var packet domain.MarketDataPacket
			if err := json.Unmarshal(message, &packet); err != nil {
				log.Printf("Failed to parse market data packet: %v", err)
				h.sendError(conn, "Invalid packet format")
				continue
			}

			// Validate packet
			if err := h.validatePacket(&packet); err != nil {
				h.sendError(conn, err.Error())
				continue
			}

			// Process the packet
			if err := h.ingestionSvc.ProcessPacket(h.ctx, &packet); err != nil {
				log.Printf("Failed to process packet: %v", err)
				h.sendError(conn, "Failed to process packet")
				continue
			}

			// Send acknowledgment
			h.sendAck(conn, packet.SequenceNum)

			// Reset read deadline
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		}
	}
}

// HandleAlertStream handles WebSocket connections for real-time alert streaming
func (h *WebSocketHandler) HandleAlertStream(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Create subscription channel
	alertChan := make(chan *domain.Alert, 100)

	// Generate subscription ID
	subID := uuid.New().String()

	// Register subscription
	h.alertSubMu.Lock()
	h.alertSubs[subID] = map[chan *domain.Alert]bool{alertChan: true}
	h.alertSubMu.Unlock()

	log.Printf("New alert stream subscription: %s", subID)

	// Start goroutine to send alerts
	h.wg.Add(1)
	go h.alertSender(conn, alertChan)

	// Handle incoming messages (for subscription control)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse control message
		var ctrlMsg ControlMessage
		if err := json.Unmarshal(message, &ctrlMsg); err != nil {
			continue
		}

		switch ctrlMsg.Action {
		case "ping":
			h.sendAlertControl(conn, "pong", subID)
		case "subscribe":
			// Already subscribed
		case "unsubscribe":
			close(alertChan)
			h.alertSubMu.Lock()
			delete(h.alertSubs, subID)
			h.alertSubMu.Unlock()
			return
		}
	}

	// Cleanup on disconnect
	close(alertChan)
	h.alertSubMu.Lock()
	if subs, ok := h.alertSubs[subID]; ok {
		delete(subs, alertChan)
		if len(subs) == 0 {
			delete(h.alertSubs, subID)
		}
	}
	h.alertSubMu.Unlock()
}

// alertSender sends alerts to the WebSocket connection
func (h *WebSocketHandler) alertSender(conn *websocket.Conn, alertChan <-chan *domain.Alert) {
	defer h.wg.Done()

	for alert := range alertChan {
		data, err := json.Marshal(alert)
		if err != nil {
			log.Printf("Failed to marshal alert: %v", err)
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Failed to send alert: %v", err)
			break
		}
	}
}

// BroadcastAlert sends an alert to all subscribed clients
func (h *WebSocketHandler) BroadcastAlert(alert *domain.Alert) {
	h.alertSubMu.RLock()
	defer h.alertSubMu.RUnlock()

	for _, subs := range h.alertSubs {
		for ch := range subs {
			select {
			case ch <- alert:
			default:
				// Channel full, skip
			}
		}
	}
}

// validatePacket validates a market data packet
func (h *WebSocketHandler) validatePacket(packet *domain.MarketDataPacket) error {
	if packet.ExchangeID == uuid.Nil {
		return &ValidationError{Field: "exchange_id", Message: "Exchange ID is required"}
	}
	if packet.Symbol == "" {
		return &ValidationError{Field: "symbol", Message: "Symbol is required"}
	}
	if packet.Timestamp.IsZero() {
		return &ValidationError{Field: "timestamp", Message: "Timestamp is required"}
	}

	// Validate trades
	for _, trade := range packet.Trades {
		if trade.TradeID == uuid.Nil {
			return &ValidationError{Field: "trade_id", Message: "Trade ID is required"}
		}
		if trade.Price.LessThanOrEqual(decimal.Zero) {
			return &ValidationError{Field: "price", Message: "Price must be positive"}
		}
		if trade.Quantity.LessThanOrEqual(decimal.Zero) {
			return &ValidationError{Field: "quantity", Message: "Quantity must be positive"}
		}
	}

	// Validate orders
	for _, order := range packet.Orders {
		if order.OrderID == uuid.Nil {
			return &ValidationError{Field: "order_id", Message: "Order ID is required"}
		}
		if order.Price.LessThanOrEqual(decimal.Zero) {
			return &ValidationError{Field: "price", Message: "Price must be positive"}
		}
	}

	return nil
}

// sendError sends an error message to the WebSocket connection
func (h *WebSocketHandler) sendError(conn *websocket.Conn, message string) {
	response := WSResponse{
		Type:    "error",
		Payload: map[string]string{"message": message},
	}

	data, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, data)
}

// sendAck sends an acknowledgment to the WebSocket connection
func (h *WebSocketHandler) sendAck(conn *websocket.Conn, sequenceNum int64) {
	response := WSResponse{
		Type: "ack",
		Payload: map[string]interface{}{
			"sequence_num": sequenceNum,
			"timestamp":    time.Now().UTC(),
		},
	}

	data, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, data)
}

// sendAlertControl sends a control message to the alert stream
func (h *WebSocketHandler) sendAlertControl(conn *websocket.Conn, action, subscriptionID string) {
	response := WSResponse{
		Type: "control",
		Payload: map[string]string{
			"action":        action,
			"subscription":  subscriptionID,
		},
	}

	data, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, data)
}

// ControlMessage represents a WebSocket control message
type ControlMessage struct {
	Action string `json:"action"`
}

// WSResponse represents a WebSocket response
type WSResponse struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Import decimal for validation
import "github.com/shopspring/decimal"
