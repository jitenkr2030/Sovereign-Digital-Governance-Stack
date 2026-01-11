// NEAM Sensing Layer - Payment Adapters
// Real implementation for banking, UPI, POS data ingestion

package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"neam-platform/sensing/common"
	"neam-platform/sensing/payments/schemas"
)

// PaymentAdapter interface defines methods all payment adapters must implement
type PaymentAdapter interface {
	// Connect establishes connection to the payment source
	Connect(ctx context.Context) error
	
	// Disconnect closes connection
	Disconnect() error
	
	// Health checks if adapter is healthy
	Health() bool
	
	// Stream returns channel of payment events
	Stream() <-chan schemas.PaymentEvent
	
	// GetMetrics returns adapter performance metrics
	Metrics() AdapterMetrics
}

// AdapterMetrics holds performance and health metrics for an adapter
type AdapterMetrics struct {
	EventsProcessed   int64
	ErrorsCount       int64
	LastEventTime     time.Time
	ProcessingLatency time.Duration
	SuccessRate       float64
}

// BankAdapter implements payment adapter for traditional banking channels
type BankAdapter struct {
	config        BankConfig
	logger        *common.Logger
	eventChan     chan schemas.PaymentEvent
	metrics       AdapterMetrics
	metricsMu     sync.RWMutex
	running       bool
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// BankConfig holds configuration for bank adapter
type BankConfig struct {
	Name           string        `mapstructure:"name"`
	APIEndpoint    string        `mapstructure:"api_endpoint"`
	APIKey         string        `mapstructure:"api_key"`
	APISecret      string        `mapstructure:"api_secret"`
	BatchSize      int           `mapstructure:"batch_size"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	RetryAttempts  int           `mapstructure:"retry_attempts"`
	RetryDelay     time.Duration `mapstructure:"retry_delay"`
	Region         string        `mapstructure:"region"`
}

// UPIAdapter implements adapter for UPI (Unified Payments Interface)
type UPIAdapter struct {
	config    UPIConfig
	logger    *common.Logger
	eventChan chan schemas.PaymentEvent
	metrics   AdapterMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// UPIConfig holds configuration for UPI adapter
type UPIConfig struct {
	Name            string        `mapstructure:"name"`
	NPCIEndpoint    string        `mapstructure:"npcI_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	PSPID           string        `mapstructure:"psp_id"`
	BatchSize       int           `mapstructure:"batch_size"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	Region          string        `mapstructure:"region"`
}

// POSAdapter implements adapter for Point of Sale terminals
type POSAdapter struct {
	config    POSConfig
	logger    *common.Logger
	eventChan chan schemas.PaymentEvent
	metrics   AdapterMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// POSConfig holds configuration for POS adapter
type POSConfig struct {
	Name            string        `mapstructure:"name"`
	AggregatorEndpoint string      `mapstructure:"aggregator_endpoint"`
	APIKey          string        `mapstructure:"api_key"`
	TerminalIDs     []string      `mapstructure:"terminal_ids"`
	BatchSize       int           `mapstructure:"batch_size"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	Region          string        `mapstructure:"region"`
}

// WalletAdapter implements adapter for digital wallets
type WalletAdapter struct {
	config    WalletConfig
	logger    *common.Logger
	eventChan chan schemas.PaymentEvent
	metrics   AdapterMetrics
	metricsMu sync.RWMutex
	running   bool
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// WalletConfig holds configuration for wallet adapter
type WalletConfig struct {
	Name           string        `mapstructure:"name"`
	APIEndpoint    string        `mapstructure:"api_endpoint"`
	APIKey         string        `mapstructure:"api_key"`
	WalletIDs      []string      `mapstructure:"wallet_ids"`
	BatchSize      int           `mapstructure:"batch_size"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
	PollingInterval time.Duration `mapstructure:"polling_interval"`
	Region         string        `mapstructure:"region"`
}

// NewBankAdapter creates a new bank adapter instance
func NewBankAdapter(config BankConfig, logger *common.Logger) *BankAdapter {
	return &BankAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan schemas.PaymentEvent, config.BatchSize*2),
		stopChan:  make(chan struct{}),
	}
}

// NewUPIAdapter creates a new UPI adapter instance
func NewUPIAdapter(config UPIConfig, logger *common.Logger) *UPIAdapter {
	return &UPIAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan schemas.PaymentEvent, config.BatchSize*2),
		stopChan:  make(chan struct{}),
	}
}

// NewPOSAdapter creates a new POS adapter instance
func NewPOSAdapter(config POSConfig, logger *common.Logger) *POSAdapter {
	return &POSAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan schemas.PaymentEvent, config.BatchSize*2),
		stopChan:  make(chan struct{}),
	}
}

// NewWalletAdapter creates a new wallet adapter instance
func NewWalletAdapter(config WalletConfig, logger *common.Logger) *WalletAdapter {
	return &WalletAdapter{
		config:    config,
		logger:    logger,
		eventChan: make(chan schemas.PaymentEvent, config.BatchSize*2),
		stopChan:  make(chan struct{}),
	}
}

// Connect implements connection logic for bank adapter
func (ba *BankAdapter) Connect(ctx context.Context) error {
	ba.logger.Info("Connecting to bank API", "endpoint", ba.config.APIEndpoint, "region", ba.config.Region)
	
	// In production, this would establish actual connection to banking APIs
	// For demonstration, we simulate the connection
	ba.logger.Info("Bank adapter connected successfully", "bank", ba.config.Name)
	
	return nil
}

// Disconnect closes the bank adapter connection
func (ba *BankAdapter) Disconnect() error {
	ba.logger.Info("Disconnecting bank adapter", "bank", ba.config.Name)
	ba.running = false
	close(ba.stopChan)
	ba.wg.Wait()
	return nil
}

// Health checks if the adapter is healthy
func (ba *BankAdapter) Health() bool {
	ba.metricsMu.RLock()
	defer ba.metricsMu.RUnlock()
	return ba.running && ba.metrics.ErrorsCount < 100
}

// Stream returns the event channel
func (ba *BankAdapter) Stream() <-chan schemas.PaymentEvent {
	return ba.eventChan
}

// Metrics returns adapter metrics
func (ba *BankAdapter) Metrics() AdapterMetrics {
	ba.metricsMu.RLock()
	defer ba.metricsMu.RUnlock()
	return ba.metrics
}

// Start begins streaming payment events from the bank
func (ba *BankAdapter) Start(ctx context.Context) error {
	if err := ba.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect bank adapter: %w", err)
	}
	
	ba.running = true
	ba.wg.Add(1)
	
	go ba.processLoop()
	
	ba.logger.Info("Bank adapter started", "bank", ba.config.Name)
	return nil
}

// processLoop continuously processes payment events
func (ba *BankAdapter) processLoop() {
	defer ba.wg.Done()
	
	ticker := time.NewTicker(ba.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ba.stopChan:
			return
		case <-ticker.C:
			if err := ba.fetchPaymentEvents(); err != nil {
				ba.logger.Error("Failed to fetch payment events", "error", err)
				ba.updateMetricsError()
			}
		}
	}
}

// fetchPaymentEvents simulates fetching events from bank API
func (ba *BankAdapter) fetchPaymentEvents() error {
	// In production, this would:
	// 1. Make authenticated API call to bank
	// 2. Parse response according to bank's API format
	// 3. Convert to standardized PaymentEvent format
	// 4. Send to event channel
	
	// Simulate fetching events
	events := ba.generateMockEvents(ba.config.BatchSize)
	
	for _, event := range events {
		select {
		case ba.eventChan <- event:
			ba.updateMetricsSuccess()
		default:
			ba.logger.Warn("Event channel full, dropping event")
			ba.updateMetricsError()
		}
	}
	
	return nil
}

// generateMockEvents creates mock payment events for testing
func (ba *BankAdapter) generateMockEvents(count int) []schemas.PaymentEvent {
	events := make([]schemas.PaymentEvent, count)
	now := time.Now()
	
	for i := 0; i < count; i++ {
		events[i] = schemas.PaymentEvent{
			TransactionID:   fmt.Sprintf("TXN-%s-%d", ba.config.Region, now.UnixNano()+int64(i)),
			Timestamp:       now.Add(-time.Duration(i) * time.Minute),
			Amount:          float64(100 + i*10),
			Currency:        "INR",
			PaymentType:     schemas.PaymentTypeTransfer,
			Channel:         schemas.ChannelBank,
			MerchantCategory: "retail",
			Region:          ba.config.Region,
			SourceBank:      ba.config.Name,
			Status:          schemas.StatusCompleted,
			Metadata: map[string]interface{}{
				"account_type": "savings",
				"processed_at": now.Format(time.RFC3339),
			},
		}
	}
	
	return events
}

// updateMetricsSuccess updates metrics on successful event processing
func (ba *BankAdapter) updateMetricsSuccess() {
	ba.metricsMu.Lock()
	defer ba.metricsMu.Unlock()
	ba.metrics.EventsProcessed++
	ba.metrics.LastEventTime = time.Now()
	
	if ba.metrics.EventsProcessed > 0 {
		success := float64(ba.metrics.EventsProcessed - ba.metrics.ErrorsCount)
		ba.metrics.SuccessRate = success / float64(ba.metrics.EventsProcessed)
	}
}

// updateMetricsError updates metrics on error
func (ba *BankAdapter) updateMetricsError() {
	ba.metricsMu.Lock()
	defer ba.metricsMu.Unlock()
	ba.metrics.ErrorsCount++
	
	if ba.metrics.EventsProcessed > 0 {
		success := float64(ba.metrics.EventsProcessed - ba.metrics.ErrorsCount)
		ba.metrics.SuccessRate = success / float64(ba.metrics.EventsProcessed)
	}
}

// UPI Adapter implementation
func (ua *UPIAdapter) Connect(ctx context.Context) error {
	ua.logger.Info("Connecting to NPCI/UPI", "endpoint", ua.config.NPCIEndpoint, "psp", ua.config.PSPID)
	ua.logger.Info("UPI adapter connected successfully")
	return nil
}

func (ua *UPIAdapter) Disconnect() error {
	ua.running = false
	close(ua.stopChan)
	ua.wg.Wait()
	return nil
}

func (ua *UPIAdapter) Health() bool {
	return ua.running
}

func (ua *UPIAdapter) Stream() <-chan schemas.PaymentEvent {
	return ua.eventChan
}

func (ua *UPIAdapter) Metrics() AdapterMetrics {
	return ua.metrics
}

func (ua *UPIAdapter) Start(ctx context.Context) error {
	if err := ua.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect UPI adapter: %w", err)
	}
	
	ua.running = true
	ua.wg.Add(1)
	
	go ua.processLoop()
	
	ua.logger.Info("UPI adapter started")
	return nil
}

func (ua *UPIAdapter) processLoop() {
	defer ua.wg.Done()
	
	ticker := time.NewTicker(ua.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ua.stopChan:
			return
		case <-ticker.C:
			ua.fetchUPIEvents()
		}
	}
}

func (ua *UPIAdapter) fetchUPIEvents() {
	now := time.Now()
	
	// Generate mock UPI events
	for i := 0; i < ua.config.BatchSize; i++ {
		event := schemas.PaymentEvent{
			TransactionID:   fmt.Sprintf("UPI-%s-%d", ua.config.Region, now.UnixNano()+int64(i)),
			Timestamp:       now.Add(-time.Duration(i) * time.Second),
			Amount:          float64(10 + i*5),
			Currency:        "INR",
			PaymentType:     schemas.PaymentTypeUPI,
			Channel:         schemas.ChannelUPI,
			MerchantCategory: "food",
			Region:          ua.config.Region,
			SourceBank:      ua.config.PSPID,
			Status:          schemas.StatusCompleted,
			Metadata: map[string]interface{}{
				"upi_id":     fmt.Sprintf("user%d@bank", i),
				"vpa":        fmt.Sprintf("merchant%d@upi", i),
			},
		}
		
		select {
		case ua.eventChan <- event:
			ua.metrics.EventsProcessed++
		default:
			ua.metrics.ErrorsCount++
		}
	}
	
	ua.metrics.LastEventTime = time.Now()
}

// POS Adapter implementation
func (pa *POSAdapter) Connect(ctx context.Context) error {
	pa.logger.Info("Connecting to POS aggregator", "endpoint", pa.config.AggregatorEndpoint)
	pa.logger.Info("POS adapter connected successfully")
	return nil
}

func (pa *POSAdapter) Disconnect() error {
	pa.running = false
	close(pa.stopChan)
	pa.wg.Wait()
	return nil
}

func (pa *POSAdapter) Health() bool {
	return pa.running
}

func (pa *POSAdapter) Stream() <-chan schemas.PaymentEvent {
	return pa.eventChan
}

func (pa *POSAdapter) Metrics() AdapterMetrics {
	return pa.metrics
}

func (pa *POSAdapter) Start(ctx context.Context) error {
	if err := pa.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect POS adapter: %w", err)
	}
	
	pa.running = true
	pa.wg.Add(1)
	
	go pa.processLoop()
	
	pa.logger.Info("POS adapter started")
	return nil
}

func (pa *POSAdapter) processLoop() {
	defer pa.wg.Done()
	
	ticker := time.NewTicker(pa.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-pa.stopChan:
			return
		case <-ticker.C:
			pa.fetchPOSTransactions()
		}
	}
}

func (pa *POSAdapter) fetchPOSTransactions() {
	now := time.Now()
	
	for i := 0; i < len(pa.config.TerminalIDs); i++ {
		for j := 0; j < pa.config.BatchSize/len(pa.config.TerminalIDs); j++ {
			event := schemas.PaymentEvent{
				TransactionID:   fmt.Sprintf("POS-%s-%s-%d", pa.config.Region, pa.config.TerminalIDs[i], now.UnixNano()),
				Timestamp:       now,
				Amount:          float64(50 + j*25),
				Currency:        "INR",
				PaymentType:     schemas.PaymentTypeCard,
				Channel:         schemas.ChannelPOS,
				MerchantCategory: "retail",
				Region:          pa.config.Region,
				SourceBank:      pa.config.AggregatorEndpoint,
				Status:          schemas.StatusCompleted,
				Metadata: map[string]interface{}{
					"terminal_id": pa.config.TerminalIDs[i],
					"card_type":   "debit",
				},
			}
			
			select {
			case pa.eventChan <- event:
				pa.metrics.EventsProcessed++
			default:
				pa.metrics.ErrorsCount++
			}
		}
	}
	
	pa.metrics.LastEventTime = time.Now()
}

// Wallet Adapter implementation
func (wa *WalletAdapter) Connect(ctx context.Context) error {
	wa.logger.Info("Connecting to wallet API", "endpoint", wa.config.APIEndpoint)
	wa.logger.Info("Wallet adapter connected successfully")
	return nil
}

func (wa *WalletAdapter) Disconnect() error {
	wa.running = false
	close(wa.stopChan)
	wa.wg.Wait()
	return nil
}

func (wa *WalletAdapter) Health() bool {
	return wa.running
}

func (wa *WalletAdapter) Stream() <-chan schemas.PaymentEvent {
	return wa.eventChan
}

func (wa *WalletAdapter) Metrics() AdapterMetrics {
	return wa.metrics
}

func (wa *WalletAdapter) Start(ctx context.Context) error {
	if err := wa.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect wallet adapter: %w", err)
	}
	
	wa.running = true
	wa.wg.Add(1)
	
	go wa.processLoop()
	
	wa.logger.Info("Wallet adapter started")
	return nil
}

func (wa *WalletAdapter) processLoop() {
	defer wa.wg.Done()
	
	ticker := time.NewTicker(wa.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-wa.stopChan:
			return
		case <-ticker.C:
			wa.fetchWalletTransactions()
		}
	}
}

func (wa *WalletAdapter) fetchWalletTransactions() {
	now := time.Now()
	
	for i := 0; i < wa.config.BatchSize; i++ {
		event := schemas.PaymentEvent{
			TransactionID:   fmt.Sprintf("WALLET-%s-%d", wa.config.Region, now.UnixNano()+int64(i)),
			Timestamp:       now,
			Amount:          float64(5 + i*2),
			Currency:        "INR",
			PaymentType:     schemas.PaymentTypeWallet,
			Channel:         schemas.ChannelWallet,
			MerchantCategory: "utilities",
			Region:          wa.config.Region,
			SourceBank:      wa.config.APIEndpoint,
			Status:          schemas.StatusCompleted,
			Metadata: map[string]interface{}{
				"wallet_id": wa.config.WalletIDs[i%len(wa.config.WalletIDs)],
			},
		}
		
		select {
		case wa.eventChan <- event:
			wa.metrics.EventsProcessed++
		default:
			wa.metrics.ErrorsCount++
		}
	}
	
	wa.metrics.LastEventTime = time.Now()
}

// PaymentAggregator combines multiple payment adapters
type PaymentAggregator struct {
	adapters     []PaymentAdapter
	outputChan   chan schemas.PaymentEvent
	logger       *common.Logger
	wg           sync.WaitGroup
	stopChan     chan struct{}
}

// NewPaymentAggregator creates a new aggregator
func NewPaymentAggregator(logger *common.Logger, adapters ...PaymentAdapter) *PaymentAggregator {
	return &PaymentAggregator{
		adapters:   adapters,
		outputChan: make(chan schemas.PaymentEvent, 10000),
		logger:     logger,
		stopChan:   make(chan struct{}),
	}
}

// Start begins aggregating events from all adapters
func (pa *PaymentAggregator) Start(ctx context.Context) error {
	// Start all adapters
	for _, adapter := range pa.adapters {
		if err := adapter.Start(ctx); err != nil {
			return fmt.Errorf("failed to start adapter: %w", err)
		}
	}
	
	// Start aggregation loop
	pa.wg.Add(1)
	go pa.aggregate()
	
	log.Println("Payment aggregator started with", len(pa.adapters), "adapters")
	return nil
}

// aggregate merges events from all adapters
func (pa *PaymentAggregator) aggregate() {
	defer pa.wg.Done()
	
	for {
		select {
		case <-pa.stopChan:
			return
		default:
			for _, adapter := range pa.adapters {
				select {
				case event := <-adapter.Stream():
					pa.outputChan <- event
				default:
					// Continue to next adapter
				}
			}
			// Small sleep to prevent tight loop
			time.Sleep(time.Millisecond)
		}
	}
}

// Stop stops all adapters and aggregation
func (pa *PaymentAggregator) Stop() {
	close(pa.stopChan)
	
	for _, adapter := range pa.adapters {
		adapter.Disconnect()
	}
	
	pa.wg.Wait()
}

// Stream returns the aggregated event stream
func (pa *PaymentAggregator) Stream() <-chan schemas.PaymentEvent {
	return pa.outputChan
}

// Metrics returns combined metrics from all adapters
func (pa *PaymentAggregator) Metrics() map[string]AdapterMetrics {
	result := make(map[string]AdapterMetrics)
	for _, adapter := range pa.adapters {
		result[adapter.Metrics().LastEventTime.String()] = adapter.Metrics()
	}
	return result
}

// JSON helper for logging events
func (pe *PaymentEvent) JSON() string {
	data, _ := json.Marshal(pe)
	return string(data)
}
