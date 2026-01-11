package adapters

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/csic/oversight/internal/core/domain"
	"github.com/csic/oversight/internal/core/ports"
	"go.uber.org/zap"
)

// HTTPServerAdapter provides the REST API for the Oversight service
type HTTPServerAdapter struct {
	router       chi.Router
	service      ports.OversightService
	eventPort    ports.OversightEventPort
	logger       *zap.Logger
}

// NewHTTPServerAdapter creates a new HTTP server adapter
func NewHTTPServerAdapter(service ports.OversightService, eventPort ports.OversightEventPort, logger *zap.Logger) *HTTPServerAdapter {
	adapter := &HTTPServerAdapter{
		router:    chi.NewRouter(),
		service:   service,
		eventPort: eventPort,
		logger:    logger,
	}
	adapter.setupMiddleware()
	adapter.setupRoutes()
	return adapter
}

// setupMiddleware configures CORS and request logging
func (a *HTTPServerAdapter) setupMiddleware() {
	// CORS configuration
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	a.router.Use(cors.Handler)
}

// setupRoutes configures the API routes
func (a *HTTPServerAdapter) setupRoutes() {
	a.router.Get("/health", a.healthCheck)
	a.router.Get("/ready", a.readinessCheck)
	
	// Exchange management
	a.router.Route("/api/v1/exchanges", func(r chi.Router) {
		r.Post("/", a.registerExchange)
		r.Get("/", a.listExchanges)
		r.Get("/{id}", a.getExchange)
		r.Get("/{id}/health", a.getExchangeHealth)
		r.Put("/{id}/status", a.updateExchangeStatus)
	})
	
	// Market data
	a.router.Route("/api/v1/market", func(r chi.Router) {
		r.Post("/trades", a.processTrade)
		r.Post("/depth", a.processMarketDepth)
		r.Get("/overview", a.getMarketOverview)
	})
	
	// Anomaly detection
	a.router.Route("/api/v1/anomalies", func(r chi.Router) {
		r.Get("/", a.getAnomalies)
		r.Get("/pending", a.getPendingAnomalies)
		r.Put("/{id}/resolve", a.resolveAnomaly)
		r.Post("/detect/wash-trading", a.detectWashTrading)
		r.Post("/detect/spoofing", a.detectSpoofing)
	})
}

// Health and readiness endpoints
func (a *HTTPServerAdapter) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "oversight",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (a *HTTPServerAdapter) readinessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
		"service": "oversight",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Exchange management handlers
func (a *HTTPServerAdapter) registerExchange(w http.ResponseWriter, r *http.Request) {
	var exchange domain.Exchange
	if err := json.NewDecoder(r.Body).Decode(&exchange); err != nil {
		a.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if err := a.service.RegisterExchange(r.Context(), &exchange); err != nil {
		a.respondError(w, http.StatusInternalServerError, "Failed to register exchange")
		return
	}
	
	a.respondJSON(w, http.StatusCreated, exchange)
}

func (a *HTTPServerAdapter) listExchanges(w http.ResponseWriter, r *http.Request) {
	exchanges, err := a.service.GetMarketOverview(r.Context())
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, "Failed to list exchanges")
		return
	}
	a.respondJSON(w, http.StatusOK, exchanges)
}

func (a *HTTPServerAdapter) getExchange(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	exchange, err := a.service.GetExchangeHealth(r.Context(), id)
	if err != nil {
		a.respondError(w, http.StatusNotFound, "Exchange not found")
		return
	}
	a.respondJSON(w, http.StatusOK, exchange)
}

func (a *HTTPServerAdapter) getExchangeHealth(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	healthScore, err := a.service.CalculateHealthScore(r.Context(), id)
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, "Failed to calculate health score")
		return
	}
	a.respondJSON(w, http.StatusOK, map[string]interface{}{
		"exchange_id": id,
		"health_score": healthScore,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (a *HTTPServerAdapter) updateExchangeStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	status := domain.ExchangeStatus(req.Status)
	if err := a.service.UpdateExchangeStatus(r.Context(), id, status); err != nil {
		a.respondError(w, http.StatusInternalServerError, "Failed to update status")
		return
	}
	a.respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// Market data handlers
func (a *HTTPServerAdapter) processTrade(w http.ResponseWriter, r *http.Request) {
	var trade domain.Trade
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		a.respondError(w, http.StatusBadRequest, "Invalid trade data")
		return
	}
	
	if err := a.service.ProcessTrade(r.Context(), &trade); err != nil {
		a.respondError(w, http.StatusInternalServerError, "Failed to process trade")
		return
	}
	a.respondJSON(w, http.StatusAccepted, map[string]string{"status": "processed"})
}

func (a *HTTPServerAdapter) processMarketDepth(w http.ResponseWriter, r *http.Request) {
	var depth domain.MarketDepth
	if err := json.NewDecoder(r.Body).Decode(&depth); err != nil {
		a.respondError(w, http.StatusBadRequest, "Invalid market depth data")
		return
	}
	
	if err := a.service.ProcessMarketDepth(r.Context(), &depth); err != nil {
		a.respondError(w, http.StatusInternalServerError, "Failed to process market depth")
		return
	}
	a.respondJSON(w, http.StatusAccepted, map[string]string{"status": "processed"})
}

func (a *HTTPServerAdapter) getMarketOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := a.service.GetMarketOverview(r.Context())
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, "Failed to get market overview")
		return
	}
	a.respondJSON(w, http.StatusOK, overview)
}

// Anomaly detection handlers
func (a *HTTPServerAdapter) getAnomalies(w http.ResponseWriter, r *http.Request) {
	exchangeID := r.URL.Query().Get("exchange_id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	
	from, _ := time.Parse(time.RFC3339, fromStr)
	to, _ := time.Parse(time.RFC3339, toStr)
	
	if from.IsZero() {
		from = time.Now().Add(-24 * time.Hour)
	}
	if to.IsZero() {
		to = time.Now()
	}
	
	anomalies, err := a.service.GetAnomalies(r.Context(), exchangeID, from, to)
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, "Failed to get anomalies")
		return
	}
	a.respondJSON(w, http.StatusOK, anomalies)
}

func (a *HTTPServerAdapter) getPendingAnomalies(w http.ResponseWriter, r *http.Request) {
	a.respondJSON(w, http.StatusOK, map[string]string{"message": "Not implemented yet"})
}

func (a *HTTPServerAdapter) resolveAnomaly(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Resolution string `json:"resolution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if err := a.service.ResolveAnomaly(r.Context(), id, req.Resolution); err != nil {
		a.respondError(w, http.StatusInternalServerError, "Failed to resolve anomaly")
		return
	}
	a.respondJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func (a *HTTPServerAdapter) detectWashTrading(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExchangeID string `json:"exchange_id"`
		Symbol     string `json:"symbol"`
		WindowSecs int    `json:"window_secs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	window := time.Duration(req.WindowSecs) * time.Second
	anomalies, err := a.service.DetectWashTrading(r.Context(), req.ExchangeID, req.Symbol, window)
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, "Detection failed")
		return
	}
	a.respondJSON(w, http.StatusOK, anomalies)
}

func (a *HTTPServerAdapter) detectSpoofing(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExchangeID string `json:"exchange_id"`
		Symbol     string `json:"symbol"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	anomalies, err := a.service.DetectSpoofing(r.Context(), req.ExchangeID, req.Symbol)
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, "Detection failed")
		return
	}
	a.respondJSON(w, http.StatusOK, anomalies)
}

// Helper methods
func (a *HTTPServerAdapter) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (a *HTTPServerAdapter) respondError(w http.ResponseWriter, status int, message string) {
	a.respondJSON(w, status, map[string]string{"error": message})
}

// GetRouter returns the chi router for serving
func (a *HTTPServerAdapter) GetRouter() chi.Router {
	return a.router
}

// GetPort returns the configured port
func (a *HTTPServerAdapter) GetPort() string {
	return ":8085"
}
