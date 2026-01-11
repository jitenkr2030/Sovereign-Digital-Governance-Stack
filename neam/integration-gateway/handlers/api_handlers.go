package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"integration-gateway/adapters"
	"integration-gateway/kafka"
	"integration-gateway/models"
	"integration-gateway/transformers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

// APIHandler handles REST API requests
type APIHandler struct {
	adapterRegistry *adapters.AdapterRegistry
	transformerReg  *transformers.TransformerRegistry
	kafkaProducer   *kafka.KafkaProducer
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(
	adapterRegistry *adapters.AdapterRegistry,
	transformerReg *transformers.TransformerRegistry,
	kafkaProducer *kafka.KafkaProducer,
) *APIHandler {
	return &APIHandler{
		adapterRegistry: adapterRegistry,
		transformerReg:  transformerReg,
		kafkaProducer:   kafkaProducer,
	}
}

// HealthCheck handles health check requests
func (h *APIHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "integration-gateway",
	})
}

// SchemaHandler returns information about supported schemas
func (h *APIHandler) SchemaHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"schemas": h.transformerReg.GetSupportedSchemas(),
	})
}

// IntegrationHandler handles incoming integration requests
func (h *APIHandler) IntegrationHandler(c *gin.Context) {
	var req models.IntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get the appropriate adapter
	adapter, err := h.adapterRegistry.GetAdapter(req.SourceSystem)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "unsupported source system",
			"details": err.Error(),
		})
		return
	}

	// Fetch data from the source system
	data, err := adapter.FetchData(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to fetch data",
			"details": err.Error(),
		})
		return
	}

	// Transform to canonical format
	transformer, err := h.transformerReg.GetTransformer(req.SchemaVersion)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "unsupported schema version",
			"details": err.Error(),
		})
		return
	}

	canonicalEvents, err := transformer.TransformBatch(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "transformation failed",
			"details": err.Error(),
		})
		return
	}

	// Publish to Kafka
	for _, event := range canonicalEvents {
		if err := h.kafkaProducer.PublishEvent(event); err != nil {
			log.Printf("Failed to publish event: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to publish events",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":           "accepted",
		"events_processed": len(canonicalEvents),
		"event_ids":        extractEventIDs(canonicalEvents),
	})
}

// BatchIntegrationHandler handles batch integration requests
func (h *APIHandler) BatchIntegrationHandler(c *gin.Context) {
	var batchReq models.BatchIntegrationRequest
	if err := c.ShouldBindJSON(&batchReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid batch request",
			"details": err.Error(),
		})
		return
	}

	results := make([]models.BatchResult, 0, len(batchReq.Requests))

	for _, req := range batchReq.Requests {
		adapter, err := h.adapterRegistry.GetAdapter(req.SourceSystem)
		if err != nil {
			results = append(results, models.BatchResult{
				RequestID: req.RequestID,
				Status:    "failed",
				Error:     err.Error(),
			})
			continue
		}

		data, err := adapter.FetchData(context.Background(), &req)
		if err != nil {
			results = append(results, models.BatchResult{
				RequestID: req.RequestID,
				Status:    "failed",
				Error:     err.Error(),
			})
			continue
		}

		transformer, err := h.transformerReg.GetTransformer(req.SchemaVersion)
		if err != nil {
			results = append(results, models.BatchResult{
				RequestID: req.RequestID,
				Status:    "failed",
				Error:     err.Error(),
			})
			continue
		}

		canonicalEvents, err := transformer.TransformBatch(data)
		if err != nil {
			results = append(results, models.BatchResult{
				RequestID: req.RequestID,
				Status:    "failed",
				Error:     err.Error(),
			})
			continue
		}

		for _, event := range canonicalEvents {
			if err := h.kafkaProducer.PublishEvent(event); err != nil {
				log.Printf("Failed to publish event in batch: %v", err)
			}
		}

		results = append(results, models.BatchResult{
			RequestID: req.RequestID,
			Status:    "accepted",
			EventIDs:  extractEventIDs(canonicalEvents),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"batch_id": uuid.New().String(),
		"results":  results,
	})
}

// OutboundHandler handles outbound data distribution
func (h *APIHandler) OutboundHandler(c *gin.Context) {
	var req models.OutboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid outbound request",
			"details": err.Error(),
		})
		return
	}

	// Get the appropriate adapter for the target system
	adapter, err := h.adapterRegistry.GetAdapter(req.TargetSystem)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "unsupported target system",
			"details": err.Error(),
		})
		return
	}

	// Transform canonical event to target format
	transformer, err := h.transformerReg.GetTransformer(req.SchemaVersion)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "unsupported schema version",
			"details": err.Error(),
		})
		return
	}

	canonicalEvent, err := transformer.Transform(req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "transformation failed",
			"details": err.Error(),
		})
		return
	}

	// Send to target system
	if err := adapter.SendData(context.Background(), canonicalEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to send data",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "delivered",
		"target":       req.TargetSystem,
		"event_id":     canonicalEvent.EventID,
		"delivered_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// MetricsHandler returns integration metrics
func (h *APIHandler) MetricsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"timestamp":          time.Now().UTC().Format(time.RFC3339),
		"events_processed":   h.kafkaProducer.GetStats().EventsPublished,
		"events_failed":      h.kafkaProducer.GetStats().EventsFailed,
		"supported_schemas":  h.transformerReg.GetSupportedSchemas(),
		"connected_systems":  h.adapterRegistry.GetConnectedSystems(),
	})
}

// Helper function to extract event IDs
func extractEventIDs(events []*models.CanonicalEvent) []string {
	ids := make([]string, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.EventID)
	}
	return ids
}

// GRPCHandler handles gRPC requests (for internal service communication)
type GRPCHandler struct {
	adapterRegistry *adapters.AdapterRegistry
	transformerReg  *transformers.TransformerRegistry
	kafkaProducer   *kafka.KafkaProducer
	grpcServer      *grpc.Server
}

// NewGRPCHandler creates a new gRPC handler
func NewGRPCHandler(
	adapterRegistry *adapters.AdapterRegistry,
	transformerReg *transformers.TransformerRegistry,
	kafkaProducer *kafka.KafkaProducer,
) *GRPCHandler {
	return &GRPCHandler{
		adapterRegistry: adapterRegistry,
		transformerReg:  transformerReg,
		kafkaProducer:   kafkaProducer,
	}
}

// IntegrationService implementation for gRPC
type IntegrationService struct {
	apiHandler *APIHandler
}

// NewIntegrationService creates a new integration service
func NewIntegrationService(apiHandler *APIHandler) *IntegrationService {
	return &IntegrationService{apiHandler: apiHandler}
}

// StreamEvents streams canonical events to clients
func (s *IntegrationService) StreamEvents(req *models.StreamRequest, stream models.IntegrationService_StreamEventsServer) error {
	// Implementation for streaming events
	return nil
}

// SendEvent handles single event submission via gRPC
func (s *IntegrationService) SendEvent(ctx context.Context, event *models.CanonicalEvent) (*models.SendResponse, error) {
	if err := s.apiHandler.kafkaProducer.PublishEvent(event); err != nil {
		return &models.SendResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &models.SendResponse{
		Success:   true,
		EventID:   event.EventID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// SendBatch handles batch event submission via gRPC
func (s *IntegrationService) SendBatch(ctx context.Context, batch *models.BatchEventRequest) (*models.BatchSendResponse, error) {
	results := make([]*models.SendResponse, 0, len(batch.Events))

	for _, event := range batch.Events {
		if err := s.apiHandler.kafkaProducer.PublishEvent(event); err != nil {
			results = append(results, &models.SendResponse{
				Success: false,
				EventID: event.EventID,
				Error:   err.Error(),
			})
		} else {
			results = append(results, &models.SendResponse{
				Success:   true,
				EventID:   event.EventID,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			})
		}
	}

	return &models.BatchSendResponse{
		TotalProcessed: len(batch.Events),
		SuccessCount:   countSuccess(results),
		Results:        results,
	}, nil
}

// Count successes in batch results
func countSuccess(results []*models.SendResponse) int {
	count := 0
	for _, r := range results {
		if r.Success {
			count++
		}
	}
	return count
}

// RESTGateway serves as the REST API gateway
type RESTGateway struct {
	apiHandler *APIHandler
	mux        *runtime.ServeMux
}

// NewRESTGateway creates a new REST gateway
func NewRESTGateway(apiHandler *APIHandler) *RESTGateway {
	return &RESTGateway{
		apiHandler: apiHandler,
		mux:        runtime.NewServeMux(),
	}
}

// SetupRoutes configures all REST routes
func (g *RESTGateway) SetupRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.GET("/health", g.apiHandler.HealthCheck)
		api.GET("/schemas", g.apiHandler.SchemaHandler)
		api.GET("/metrics", g.apiHandler.MetricsHandler)

		api.POST("/integrate", g.apiHandler.IntegrationHandler)
		api.POST("/integrate/batch", g.apiHandler.BatchIntegrationHandler)
		api.POST("/outbound", g.apiHandler.OutboundHandler)
	}
}

// RequestLogger middleware logs all incoming requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[%s] %s %s - %d - %v", time.Now().Format(time.RFC3339), method, path, status, latency)
	}
}

// ErrorHandler middleware handles panics and errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}

// BodyLogger middleware logs request bodies for debugging
func BodyLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(io.LimitReader(io.MultiReader(
				io.LimitReader(io.LimitReader(
					// Note: In a real implementation, proper body restoration would be used
				), 1024),
			), 1024))
			if len(bodyBytes) > 0 {
				log.Printf("Request body: %s", string(bodyBytes[:min(1024, len(bodyBytes))]))
			}
		}
		c.Next()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// WebhookHandler handles webhooks from external systems
func (h *APIHandler) WebhookHandler(c *gin.Context) {
	sourceSystem := c.Param("source")
	if sourceSystem == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source system not specified"})
		return
	}

	// Read the webhook payload
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Parse based on content type
	var data interface{}
	contentType := c.ContentType()

	switch {
	case contentType == "application/json" || contentType == "":
		if err := json.Unmarshal(body, &data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
			return
		}
	default:
		// For other content types, treat as raw string
		data = string(body)
	}

	// Get the appropriate adapter
	adapter, err := h.adapterRegistry.GetAdapter(sourceSystem)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "unsupported source system",
			"details": err.Error(),
		})
		return
	}

	// Transform to canonical format
	transformer, err := h.transformerReg.GetTransformer(fmt.Sprintf("%s-webhook-v1.0", sourceSystem))
	if err != nil {
		// Fallback to default transformer
		transformer, err = h.transformerReg.GetTransformer("legacy-csv-v1.0")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no suitable transformer found"})
			return
		}
	}

	canonicalEvent, err := transformer.Transform(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "transformation failed",
			"details": err.Error(),
		})
		return
	}

	// Publish to Kafka
	if err := h.kafkaProducer.PublishEvent(canonicalEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to publish event",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "received",
		"event_id": canonicalEvent.EventID,
	})
}
