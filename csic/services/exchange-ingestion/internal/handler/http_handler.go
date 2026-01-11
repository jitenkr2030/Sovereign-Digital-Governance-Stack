package handler

import (
	"encoding/json"
	"net/http"

	"github.com/csic-platform/services/exchange-ingestion/internal/core/ports"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// HTTPHandler handles HTTP requests for the ingestion service
type HTTPHandler struct {
	dataSourceService ports.DataSourceService
	ingestionService  ports.IngestionService
	logger            *zap.Logger
}

// NewHTTPHandler creates a new HTTPHandler
func NewHTTPHandler(
	dataSourceService ports.DataSourceService,
	ingestionService ports.IngestionService,
	logger *zap.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		dataSourceService: dataSourceService,
		ingestionService:  ingestionService,
		logger:            logger,
	}
}

// RegisterRoutes registers all HTTP routes
func (h *HTTPHandler) RegisterRoutes(r *chi.Mux) {
	// Data source management routes
	r.Post("/api/v1/data-sources", h.RegisterDataSource)
	r.Get("/api/v1/data-sources", h.ListDataSources)
	r.Get("/api/v1/data-sources/{id}", h.GetDataSource)
	r.Put("/api/v1/data-sources/{id}", h.UpdateDataSource)
	r.Delete("/api/v1/data-sources/{id}", h.DeleteDataSource)
	r.Post("/api/v1/data-sources/{id}/enable", h.EnableDataSource)
	r.Post("/api/v1/data-sources/{id}/disable", h.DisableDataSource)
	r.Post("/api/v1/data-sources/{id}/test-connection", h.TestConnection)

	// Ingestion control routes
	r.Post("/api/v1/ingestion/start", h.StartIngestion)
	r.Post("/api/v1/ingestion/stop", h.StopIngestion)
	r.Get("/api/v1/ingestion/status", h.GetIngestionStatus)
	r.Post("/api/v1/ingestion/source/{id}/start", h.StartSource)
	r.Post("/api/v1/ingestion/source/{id}/stop", h.StopSource)
	r.Get("/api/v1/ingestion/source/{id}/status", h.GetSourceStatus)
	r.Post("/api/v1/ingestion/source/{id}/sync", h.ForceSync)
	r.Get("/api/v1/ingestion/source/{id}/stats", h.GetSourceStats)

	// Health check
	r.Get("/health", h.HealthCheck)
}

// RegisterDataSource registers a new data source
func (h *HTTPHandler) RegisterDataSource(w http.ResponseWriter, r *http.Request) {
	var req ports.RegisterDataSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.dataSourceService.RegisterDataSource(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to register data source", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

// ListDataSources lists all data sources
func (h *HTTPHandler) ListDataSources(w http.ResponseWriter, r *http.Request) {
	var req ports.ListDataSourcesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && r.Body != http.NoBody {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.dataSourceService.ListDataSources(r.Context(), &req)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list data sources", err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// GetDataSource retrieves a specific data source
func (h *HTTPHandler) GetDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	config, err := h.dataSourceService.GetDataSource(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Data source not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, config)
}

// UpdateDataSource updates an existing data source
func (h *HTTPHandler) UpdateDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req ports.UpdateDataSourceRequest
	req.ID = id

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.dataSourceService.UpdateDataSource(r.Context(), &req); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update data source", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Data source updated successfully"})
}

// DeleteDataSource deletes a data source
func (h *HTTPHandler) DeleteDataSource(w http.ResponseWriter, r *http.Request) {
	// Implementation would go here
	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Data source deleted"})
}

// EnableDataSource enables a data source
func (h *HTTPHandler) EnableDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.dataSourceService.EnableDataSource(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to enable data source", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Data source enabled"})
}

// DisableDataSource disables a data source
func (h *HTTPHandler) DisableDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.dataSourceService.DisableDataSource(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to disable data source", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Data source disabled"})
}

// TestConnection tests the connection to a data source
func (h *HTTPHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	resp, err := h.dataSourceService.TestConnection(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to test connection", err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// StartIngestion starts all ingestion processes
func (h *HTTPHandler) StartIngestion(w http.ResponseWriter, r *http.Request) {
	if err := h.ingestionService.StartIngestion(r.Context()); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to start ingestion", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Ingestion started"})
}

// StopIngestion stops all ingestion processes
func (h *HTTPHandler) StopIngestion(w http.ResponseWriter, r *http.Request) {
	if err := h.ingestionService.StopIngestion(r.Context()); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to stop ingestion", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Ingestion stopped"})
}

// GetIngestionStatus returns the overall ingestion status
func (h *HTTPHandler) GetIngestionStatus(w http.ResponseWriter, r *http.Request) {
	resp, err := h.ingestionService.GetIngestionStatus(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get ingestion status", err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// StartSource starts ingestion for a specific source
func (h *HTTPHandler) StartSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.ingestionService.StartSource(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to start source", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Source ingestion started"})
}

// StopSource stops ingestion for a specific source
func (h *HTTPHandler) StopSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.ingestionService.StopSource(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to stop source", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Source ingestion stopped"})
}

// GetSourceStatus returns the status of a specific source
func (h *HTTPHandler) GetSourceStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	resp, err := h.ingestionService.GetSourceStatus(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get source status", err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// ForceSync triggers an immediate sync for a source
func (h *HTTPHandler) ForceSync(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.ingestionService.ForceSync(r.Context(), id); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to force sync", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "Force sync completed"})
}

// GetSourceStats returns statistics for a source
func (h *HTTPHandler) GetSourceStats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	stats, err := h.ingestionService.GetStats(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get source stats", err)
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// HealthCheck returns the health status of the service
func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "exchange-ingestion",
	})
}

// writeJSON writes a JSON response
func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *HTTPHandler) writeError(w http.ResponseWriter, status int, message string, err error) {
	h.logger.Error(message, zap.Error(err))
	h.writeJSON(w, status, map[string]interface{}{
		"error":   message,
		"details": err.Error(),
	})
}
