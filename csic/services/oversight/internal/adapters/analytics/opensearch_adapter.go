package analytics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/csic/oversight/internal/core/domain"
	"github.com/csic/oversight/internal/core/ports"

	"go.uber.org/zap"
)

// OpenSearchClient manages connections to OpenSearch
type OpenSearchClient struct {
	baseURL     string
	indexPrefix string
	username    string
	password    string
	httpClient  *http.Client
	logger      *zap.Logger
}

// OpenSearchConfig holds OpenSearch configuration
type OpenSearchConfig struct {
	Addresses    []string `mapstructure:"addresses"`
	IndexPrefix  string   `mapstructure:"index_prefix"`
	TradeIndex   string   `mapstructure:"trade_index"`
	Username     string   `mapstructure:"username"`
	Password     string   `mapstructure:"password"`
	MaxRetries   int      `mapstructure:"max_retries"`
	TransportTimeout int   `mapstructure:"transport_timeout"`
}

// NewOpenSearchClient creates a new OpenSearch client
func NewOpenSearchClient(config OpenSearchConfig, logger *zap.Logger) (*OpenSearchClient, error) {
	if len(config.Addresses) == 0 {
		return nil, fmt.Errorf("OpenSearch addresses are required")
	}

	client := &OpenSearchClient{
		baseURL:     config.Addresses[0],
		indexPrefix: config.IndexPrefix,
		username:    config.Username,
		password:    config.Password,
		httpClient: &http.Client{
			Timeout: time.Duration(config.TransportTimeout) * time.Second,
		},
		logger: logger,
	}

	// Verify connection
	if err := client.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to OpenSearch: %w", err)
	}

	// Ensure index exists
	if err := client.ensureIndex(context.Background(), config.TradeIndex); err != nil {
		return nil, fmt.Errorf("failed to ensure trade index: %w", err)
	}

	return client, nil
}

// Ping checks connectivity to OpenSearch
func (c *OpenSearchClient) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return err
	}

	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("OpenSearch ping failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenSearch returned status: %d", resp.StatusCode)
	}

	return nil
}

// ensureIndex creates the index if it doesn't exist
func (c *OpenSearchClient) ensureIndex(ctx context.Context, indexName string) error {
	fullIndexName := c.getIndexName(indexName)

	// Check if index exists
	checkURL := fmt.Sprintf("%s/%s", c.baseURL, fullIndexName)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, checkURL, nil)
	if err != nil {
		return err
	}

	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil // Index already exists
	}

	// Create index with mapping
	mapping := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"trade_id":       map[string]string{"type": "keyword"},
				"exchange_id":    map[string]string{"type": "keyword"},
				"trading_pair":   map[string]string{"type": "keyword"},
				"price":          map[string]string{"type": "double"},
				"volume":         map[string]string{"type": "double"},
				"quote_volume":   map[string]string{"type": "double"},
				"buyer_user_id":  map[string]string{"type": "keyword"},
				"seller_user_id": map[string]string{"type": "keyword"},
				"timestamp":      map[string]string{"type": "date"},
				"received_at":    map[string]string{"type": "date"},
			},
		},
	}

	body, err := json.Marshal(mapping)
	if err != nil {
		return err
	}

	createReq, err := http.NewRequestWithContext(ctx, http.MethodPut, checkURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	createReq.Header.Set("Content-Type", "application/json")
	if c.username != "" && c.password != "" {
		createReq.SetBasicAuth(c.username, c.password)
	}

	createResp, err := c.httpClient.Do(createReq)
	if err != nil {
		return err
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusOK && createResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(createResp.Body)
		return fmt.Errorf("failed to create index: %s", string(respBody))
	}

	c.logger.Info("OpenSearch index created", zap.String("index", fullIndexName))
	return nil
}

// getIndexName returns the full index name with prefix
func (c *OpenSearchClient) getIndexName(indexName string) string {
	if c.indexPrefix != "" {
		return c.indexPrefix + "_" + indexName
	}
	return indexName
}

// TradeAnalyticsEngine implements the TradeAnalyticsEngine interface
type TradeAnalyticsEngine struct {
	client *OpenSearchClient
	logger *zap.Logger
}

// NewTradeAnalyticsEngine creates a new TradeAnalyticsEngine
func NewTradeAnalyticsEngine(client *OpenSearchClient, logger *zap.Logger) *TradeAnalyticsEngine {
	return &TradeAnalyticsEngine{
		client: client,
		logger: logger,
	}
}

// IndexTrade indexes a trade event
func (e *TradeAnalyticsEngine) IndexTrade(ctx context.Context, trade domain.TradeEvent) error {
	doc := map[string]interface{}{
		"trade_id":       trade.TradeID,
		"exchange_id":    trade.ExchangeID,
		"trading_pair":   trade.TradingPair,
		"price":          trade.Price,
		"volume":         trade.Volume,
		"quote_volume":   trade.QuoteVolume,
		"buyer_user_id":  trade.BuyerUserID,
		"seller_user_id": trade.SellerUserID,
		"buyer_order_id": trade.BuyerOrderID,
		"seller_order_id": trade.SellerOrderID,
		"timestamp":      trade.Timestamp.Format(time.RFC3339),
		"received_at":    trade.ReceivedAt.Format(time.RFC3339),
		"is_maker":       trade.IsMaker,
		"fee_currency":   trade.FeeCurrency,
		"fee_amount":     trade.FeeAmount,
	}

	return e.indexDocument(ctx, "trades", trade.TradeID, doc)
}

// IndexTradeBatch indexes multiple trade events
func (e *TradeAnalyticsEngine) IndexTradeBatch(ctx context.Context, trades []domain.TradeEvent) error {
	if len(trades) == 0 {
		return nil
	}

	// Use bulk API for better performance
	var buf bytes.Buffer

	for _, trade := range trades {
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": e.client.getIndexName("trades"),
				"_id":    trade.TradeID,
			},
		}

		actionBytes, err := json.Marshal(action)
		if err != nil {
			e.logger.Warn("Failed to marshal action", zap.Error(err))
			continue
		}
		buf.Write(actionBytes)
		buf.WriteByte('\n')

		doc := map[string]interface{}{
			"trade_id":        trade.TradeID,
			"exchange_id":     trade.ExchangeID,
			"trading_pair":    trade.TradingPair,
			"price":           trade.Price,
			"volume":          trade.Volume,
			"quote_volume":    trade.QuoteVolume,
			"buyer_user_id":   trade.BuyerUserID,
			"seller_user_id":  trade.SellerUserID,
			"timestamp":       trade.Timestamp.Format(time.RFC3339),
			"received_at":     trade.ReceivedAt.Format(time.RFC3339),
		}

		docBytes, err := json.Marshal(doc)
		if err != nil {
			e.logger.Warn("Failed to marshal document", zap.Error(err))
			continue
		}
		buf.Write(docBytes)
		buf.WriteByte('\n')
	}

	return e.bulkIndex(ctx, &buf)
}

// FindWashTradingPatterns detects potential wash trading patterns
func (e *TradeAnalyticsEngine) FindWashTradingPatterns(ctx context.Context, query ports.PatternQuery) ([]domain.TradeEvent, error) {
	if query.Threshold <= 0 {
		query.Threshold = 5.0
	}
	
	// Return empty result for now - pattern detection requires complex aggregation logic
	return []domain.TradeEvent{}, nil
}

// FindVolumeAnomalies detects unusual volume patterns
func (e *TradeAnalyticsEngine) FindVolumeAnomalies(ctx context.Context, query ports.PatternQuery) ([]ports.VolumeAnomaly, error) {
	// Return empty result for now - anomaly detection requires complex aggregation logic
	return []ports.VolumeAnomaly{}, nil
}

// CalculateVolumeStats calculates volume statistics for a time period
func (e *TradeAnalyticsEngine) CalculateVolumeStats(ctx context.Context, exchangeID, tradingPair string, start, end time.Time) (*ports.VolumeStats, error) {
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"exchange_id": exchangeID,
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"trading_pair": tradingPair,
						},
					},
					map[string]interface{}{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": start.Format(time.RFC3339),
								"lte": end.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
		"size": 0,
		"aggs": map[string]interface{}{
			"total_volume": map[string]interface{}{
				"sum": map[string]interface{}{
					"field": "volume",
				},
			},
			"trade_count": map[string]interface{}{
				"value_count": map[string]interface{}{
					"field": "trade_id",
				},
			},
		},
	}

	result, err := e.executeSearch(ctx, "trades", esQuery)
	if err != nil {
		return nil, err
	}

	stats := &ports.VolumeStats{
		ExchangeID:  exchangeID,
		TradingPair: tradingPair,
		PeriodStart: start,
		PeriodEnd:   end,
	}

	// Parse aggregations from response
	if agg, ok := result[0]["aggregations"]; ok {
		if totalVol, ok := agg.(map[string]interface{})["total_volume"]; ok {
			if sum, ok := totalVol.(map[string]interface{})["value"]; ok {
				if val, ok := sum.(float64); ok {
					stats.TotalVolume = val
				}
			}
		}
		if count, ok := agg.(map[string]interface{})["trade_count"]; ok {
			if val, ok := count.(map[string]interface{})["value"]; ok {
				if c, ok := val.(float64); ok {
					stats.TradeCount = int64(c)
					if c > 0 {
						stats.AvgVolume = stats.TotalVolume / c
					}
				}
			}
		}
	}

	return stats, nil
}

// GetTradeCount retrieves trade count for a query
func (e *TradeAnalyticsEngine) GetTradeCount(ctx context.Context, query ports.TradeQuery) (int64, error) {
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{},
			},
		},
		"size": 0,
	}

	must := esQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]interface{})

	if query.ExchangeID != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"exchange_id": query.ExchangeID,
			},
		})
	}

	if query.TradingPair != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"trading_pair": query.TradingPair,
			},
		})
	}

	if query.StartTime != nil && query.EndTime != nil {
		must = append(must, map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"gte": query.StartTime.Format(time.RFC3339),
					"lte": query.EndTime.Format(time.RFC3339),
				},
			},
		})
	}

	result, err := e.executeSearch(ctx, "trades", esQuery)
	if err != nil {
		return 0, err
	}

	// Extract count from hits
	if len(result) > 0 {
		if hits, ok := result[0]["hits"]; ok {
			if total, ok := hits.(map[string]interface{})["total"]; ok {
				if value, ok := total.(map[string]interface{})["value"]; ok {
					if count, ok := value.(float64); ok {
						return int64(count), nil
					}
				}
			}
		}
	}

	return 0, nil
}

// GetAggregatedVolume retrieves aggregated volume data
func (e *TradeAnalyticsEngine) GetAggregatedVolume(ctx context.Context, query ports.AggregationQuery) ([]ports.AggregationResult, error) {
	// Return empty result for now - aggregation requires executing the query
	return []ports.AggregationResult{}, nil
}

// DetectPriceManipulation detects potential price manipulation
func (e *TradeAnalyticsEngine) DetectPriceManipulation(ctx context.Context, query ports.PatternQuery) ([]ports.PriceManipulationEvent, error) {
	// Implementation would use more complex queries to detect manipulation patterns
	return []ports.PriceManipulationEvent{}, nil // Simplified
}

// indexDocument indexes a single document
func (e *TradeAnalyticsEngine) indexDocument(ctx context.Context, indexName, docID string, doc map[string]interface{}) error {
	url := fmt.Sprintf("%s/%s/_doc/%s", e.client.baseURL, e.client.getIndexName(indexName), docID)

	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if e.client.username != "" && e.client.password != "" {
		req.SetBasicAuth(e.client.username, e.client.password)
	}

	resp, err := e.client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to index document: %s", string(respBody))
	}

	return nil
}

// bulkIndex performs bulk indexing
func (e *TradeAnalyticsEngine) bulkIndex(ctx context.Context, buf *bytes.Buffer) error {
	url := fmt.Sprintf("%s/_bulk", e.client.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-ndjson")
	if e.client.username != "" && e.client.password != "" {
		req.SetBasicAuth(e.client.username, e.client.password)
	}

	resp, err := e.client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bulk index failed: %s", string(respBody))
	}

	return nil
}

// executeSearch executes a search query
func (e *TradeAnalyticsEngine) executeSearch(ctx context.Context, indexName string, query map[string]interface{}) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s/_search", e.client.baseURL, e.client.getIndexName(indexName))

	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if e.client.username != "" && e.client.password != "" {
		req.SetBasicAuth(e.client.username, e.client.password)
	}

	resp, err := e.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s", string(respBody))
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		results = append(results, hit.Source)
	}

	return results, nil
}

// Helper function to format duration for OpenSearch
func formatDuration(d time.Duration) string {
	if d >= time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d >= time.Second {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dms", int(d.Milliseconds()))
}

// Helper function to extract value from nested map
func getFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0
}

// Helper function to check if string contains
func containsString(s string, list []string) bool {
	for _, item := range list {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}
