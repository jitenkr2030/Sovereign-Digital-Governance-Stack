package service

import (
	"context"
	"testing"
	"time"

	"github.com/csic/surveillance/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMarketRepository is a mock implementation of MarketRepository
type MockMarketRepository struct {
	mock.Mock
}

func (m *MockMarketRepository) StoreEvent(ctx context.Context, event *domain.MarketEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockMarketRepository) StoreEvents(ctx context.Context, events []domain.MarketEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockMarketRepository) GetEventByID(ctx context.Context, id uuid.UUID) (*domain.MarketEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MarketEvent), args.Error(1)
}

func (m *MockMarketRepository) GetEventsBySymbol(ctx context.Context, symbol string, startTime, endTime time.Time, limit int) ([]domain.MarketEvent, error) {
	args := m.Called(ctx, symbol, startTime, endTime, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MarketEvent), args.Error(1)
}

func (m *MockMarketRepository) GetEventsByExchange(ctx context.Context, exchangeID uuid.UUID, startTime, endTime time.Time, limit int) ([]domain.MarketEvent, error) {
	args := m.Called(ctx, exchangeID, startTime, endTime, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MarketEvent), args.Error(1)
}

func (m *MockMarketRepository) GetRecentEvents(ctx context.Context, exchangeID uuid.UUID, symbol string, duration time.Duration) ([]domain.MarketEvent, error) {
	args := m.Called(ctx, exchangeID, symbol, duration)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MarketEvent), args.Error(1)
}

func (m *MockMarketRepository) GetTradesByAccount(ctx context.Context, accountID string, startTime, endTime time.Time, limit int) ([]domain.MarketEvent, error) {
	args := m.Called(ctx, accountID, startTime, endTime, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MarketEvent), args.Error(1)
}

func (m *MockMarketRepository) GetTradesByTimeWindow(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]domain.MarketEvent, error) {
	args := m.Called(ctx, exchangeID, symbol, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MarketEvent), args.Error(1)
}

func (m *MockMarketRepository) GetRelatedTrades(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time, accountIDs []string) ([]domain.MarketEvent, error) {
	args := m.Called(ctx, exchangeID, symbol, startTime, endTime, accountIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MarketEvent), args.Error(1)
}

func (m *MockMarketRepository) GetOrdersByAccount(ctx context.Context, accountID string, startTime, endTime time.Time) ([]domain.MarketEvent, error) {
	args := m.Called(ctx, accountID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MarketEvent), args.Error(1)
}

func (m *MockMarketRepository) GetCancelledOrders(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]domain.MarketEvent, error) {
	args := m.Called(ctx, exchangeID, symbol, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MarketEvent), args.Error(1)
}

func (m *MockMarketRepository) GetOrderLifetimes(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) ([]OrderLifetime, error) {
	args := m.Called(ctx, exchangeID, symbol, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]OrderLifetime), args.Error(1)
}

func (m *MockMarketRepository) GetMarketSummary(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time) (*domain.MarketSummary, error) {
	args := m.Called(ctx, exchangeID, symbol, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MarketSummary), args.Error(1)
}

func (m *MockMarketRepository) GetMarketStats(ctx context.Context, exchangeID uuid.UUID, symbol string, duration time.Duration) (*domain.MarketStats, error) {
	args := m.Called(ctx, exchangeID, symbol, duration)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MarketStats), args.Error(1)
}

func (m *MockMarketRepository) GetPriceHistory(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time, interval string) ([]domain.PricePoint, error) {
	args := m.Called(ctx, exchangeID, symbol, startTime, endTime, interval)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PricePoint), args.Error(1)
}

func (m *MockMarketRepository) GetVolumeHistory(ctx context.Context, exchangeID uuid.UUID, symbol string, startTime, endTime time.Time, interval string) ([]domain.VolumePoint, error) {
	args := m.Called(ctx, exchangeID, symbol, startTime, endTime, interval)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.VolumePoint), args.Error(1)
}

func (m *MockMarketRepository) GetGlobalAveragePrice(ctx context.Context, symbol string, startTime, endTime time.Time) (decimal.Decimal, error) {
	args := m.Called(ctx, symbol, startTime, endTime)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockMarketRepository) StoreOrderBook(ctx context.Context, snapshot *domain.OrderBookSnapshot) error {
	args := m.Called(ctx, snapshot)
	return args.Error(0)
}

func (m *MockMarketRepository) GetOrderBookSnapshot(ctx context.Context, id uuid.UUID) (*domain.OrderBookSnapshot, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrderBookSnapshot), args.Error(1)
}

func (m *MockMarketRepository) BulkStoreEvents(ctx context.Context, events []domain.MarketEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

// MockAlertRepository is a mock implementation of AlertRepository
type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) Create(ctx context.Context, alert *domain.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Alert, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Alert), args.Error(1)
}

func (m *MockAlertRepository) Update(ctx context.Context, alert *domain.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAlertRepository) List(ctx context.Context, filter domain.AlertFilter, limit, offset int) ([]domain.Alert, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Alert), args.Error(1)
}

func (m *MockAlertRepository) Count(ctx context.Context, filter domain.AlertFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAlertRepository) GetByStatus(ctx context.Context, status domain.AlertStatus) ([]domain.Alert, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetByExchange(ctx context.Context, exchangeID uuid.UUID, limit, offset int) ([]domain.Alert, error) {
	args := m.Called(ctx, exchangeID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetBySymbol(ctx context.Context, symbol string, limit, offset int) ([]domain.Alert, error) {
	args := m.Called(ctx, symbol, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MarketEvent), args.Error(1)
}

func (m *MockAlertRepository) GetOpenAlerts(ctx context.Context) ([]domain.Alert, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetRecentAlerts(ctx context.Context, since time.Time) ([]domain.Alert, error) {
	args := m.Called(ctx, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetStatistics(ctx context.Context, startTime, endTime time.Time) (*domain.AlertStatistics, error) {
	args := m.Called(ctx, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AlertStatistics), args.Error(1)
}

func (m *MockAlertRepository) GetAlertsByType(ctx context.Context, alertType domain.AlertType, startTime, endTime time.Time) ([]domain.Alert, error) {
	args := m.Called(ctx, alertType, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Alert), args.Error(1)
}

func (m *MockAlertRepository) GetAlertTrend(ctx context.Context, interval string, startTime, endTime time.Time) ([]domain.AlertTrendPoint, error) {
	args := m.Called(ctx, interval, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.AlertTrendPoint), args.Error(1)
}

func (m *MockAlertRepository) BulkCreate(ctx context.Context, alerts []domain.Alert) error {
	args := m.Called(ctx, alerts)
	return args.Error(0)
}

func (m *MockAlertRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status domain.AlertStatus) error {
	args := m.Called(ctx, ids, status)
	return args.Error(0)
}

// Test helper functions
func createTestTrade(exchangeID, symbol, userID string, price, quantity decimal.Decimal, direction domain.MarketEventDirection, timestamp time.Time) *domain.MarketEvent {
	orderID := uuid.New()
	return &domain.MarketEvent{
		ID:         uuid.New(),
		ExchangeID: exchangeID,
		Symbol:     symbol,
		EventType:  domain.MarketEventTypeTrade,
		OrderID:    &orderID,
		UserID:     &userID,
		Price:      price,
		Quantity:   quantity,
		Direction:  direction,
		Status:     domain.MarketEventStatusFilled,
		Timestamp:  timestamp,
		ReceivedAt: time.Now(),
	}
}

func createTestOrder(exchangeID, symbol, userID string, price, quantity decimal.Decimal, direction domain.MarketEventDirection, status domain.MarketEventStatus, timestamp time.Time) *domain.MarketEvent {
	orderID := uuid.New()
	return &domain.MarketEvent{
		ID:            uuid.New(),
		ExchangeID:    exchangeID,
		Symbol:        symbol,
		EventType:     domain.MarketEventTypeOrder,
		OrderID:       &orderID,
		UserID:        &userID,
		Price:         price,
		Quantity:      quantity,
		Direction:     direction,
		Status:        status,
		Timestamp:     timestamp,
		ReceivedAt:    time.Now(),
	}
}

// Analysis Service Tests

func TestDetectWashTrading_NoTrades(t *testing.T) {
	mockMarketRepo := new(MockMarketRepository)
	mockAlertRepo := new(MockAlertRepository)

	analysisSvc := NewAnalysisService(mockMarketRepo, mockAlertRepo)

	ctx := context.Background()
	exchangeID := uuid.New()
	symbol := "BTC/USD"
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()

	mockMarketRepo.On("GetTradesByTimeWindow", ctx, exchangeID, symbol, startTime, endTime).Return([]domain.MarketEvent{}, nil)

	candidates, err := analysisSvc.DetectWashTrading(ctx, exchangeID, symbol, startTime, endTime)

	assert.NoError(t, err)
	assert.Empty(t, candidates)
	mockMarketRepo.AssertExpectations(t)
}

func TestDetectWashTrading_SingleTrade(t *testing.T) {
	mockMarketRepo := new(MockMarketRepository)
	mockAlertRepo := new(MockAlertRepository)

	analysisSvc := NewAnalysisService(mockMarketRepo, mockAlertRepo)

	ctx := context.Background()
	exchangeID := uuid.New()
	symbol := "BTC/USD"
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()

	trade := createTestTrade(exchangeID.String(), symbol, "user1", decimal.NewFromFloat(50000), decimal.NewFromFloat(1), domain.MarketEventDirectionBuy, startTime)

	mockMarketRepo.On("GetTradesByTimeWindow", ctx, exchangeID, symbol, startTime, endTime).Return([]domain.MarketEvent{*trade}, nil)

	candidates, err := analysisSvc.DetectWashTrading(ctx, exchangeID, symbol, startTime, endTime)

	assert.NoError(t, err)
	assert.Empty(t, candidates)
	mockMarketRepo.AssertExpectations(t)
}

func TestDetectWashTrading_SelfTrading(t *testing.T) {
	mockMarketRepo := new(MockMarketRepository)
	mockAlertRepo := new(MockAlertRepository)

	analysisSvc := NewAnalysisService(mockMarketRepo, mockAlertRepo)

	ctx := context.Background()
	exchangeID := uuid.New()
	symbol := "BTC/USD"
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()

	// Create two trades by the same user in opposite directions with similar prices
	buyTime := startTime.Add(30 * time.Second)
	sellTime := buyTime.Add(10 * time.Second)

	buyTrade := createTestTrade(exchangeID.String(), symbol, "user1", decimal.NewFromFloat(50000), decimal.NewFromFloat(1), domain.MarketEventDirectionBuy, buyTime)
	sellTrade := createTestTrade(exchangeID.String(), symbol, "user1", decimal.NewFromFloat(50001), decimal.NewFromFloat(1), domain.MarketEventDirectionSell, sellTime)

	mockMarketRepo.On("GetTradesByTimeWindow", ctx, exchangeID, symbol, startTime, endTime).Return([]domain.MarketEvent{*buyTrade, *sellTrade}, nil)

	candidates, err := analysisSvc.DetectWashTrading(ctx, exchangeID, symbol, startTime, endTime)

	assert.NoError(t, err)
	assert.Len(t, candidates, 1)
	assert.Equal(t, domain.AlertTypeWashTrading.String(), candidates[0].AccountIDs[0])
	mockMarketRepo.AssertExpectations(t)
}

func TestDetectSpoofing_NoOrders(t *testing.T) {
	mockMarketRepo := new(MockMarketRepository)
	mockAlertRepo := new(MockAlertRepository)

	analysisSvc := NewAnalysisService(mockMarketRepo, mockAlertRepo)

	ctx := context.Background()
	exchangeID := uuid.New()
	symbol := "BTC/USD"
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()

	mockMarketRepo.On("GetCancelledOrders", ctx, exchangeID, symbol, startTime, endTime).Return([]domain.MarketEvent{}, nil)
	mockMarketRepo.On("GetOrderLifetimes", ctx, exchangeID, symbol, startTime, endTime).Return([]OrderLifetime{}, nil)

	indicators, err := analysisSvc.DetectSpoofing(ctx, exchangeID, symbol, startTime, endTime)

	assert.NoError(t, err)
	assert.Empty(t, indicators)
	mockMarketRepo.AssertExpectations(t)
}

func TestCalculateWashTradeConfidence(t *testing.T) {
	mockMarketRepo := new(MockMarketRepository)
	mockAlertRepo := new(MockAlertRepository)

	analysisSvc := NewAnalysisService(mockMarketRepo, mockAlertRepo)

	trade1 := createTestTrade("uuid", "BTC/USD", "user1", decimal.NewFromFloat(50000), decimal.NewFromFloat(1), domain.MarketEventDirectionBuy, time.Now())
	trade2 := createTestTrade("uuid", "BTC/USD", "user1", decimal.NewFromFloat(50001), decimal.NewFromFloat(1), domain.MarketEventDirectionSell, trade1.Timestamp.Add(5*time.Second))

	priceDeviation := decimal.NewFromFloat(0.0001)
	qtySimilarity := decimal.NewFromFloat(1.0)

	confidence := analysisSvc.calculateWashTradeConfidence(*trade1, *trade2, priceDeviation, qtySimilarity)

	assert.Greater(t, confidence, 0.5)
	assert.LessOrEqual(t, confidence, 1.0)
}

func TestCalculateSpoofingConfidence(t *testing.T) {
	mockMarketRepo := new(MockMarketRepository)
	mockAlertRepo := new(MockAlertRepository)

	analysisSvc := NewAnalysisService(mockMarketRepo, mockAlertRepo)

	lifetime := OrderLifetime{
		OrderID:      uuid.New(),
		AccountID:    "user1",
		InitialQty:   decimal.NewFromFloat(10),
		FilledQty:    decimal.NewFromFloat(1),
		CancelledQty: decimal.NewFromFloat(9),
		Lifetime:     60 * time.Second,
	}
	cancellationRate := decimal.NewFromFloat(0.9)

	confidence := analysisSvc.calculateSpoofingConfidence(lifetime, cancellationRate)

	assert.Greater(t, confidence, 0.7)
	assert.LessOrEqual(t, confidence, 1.0)
}

func TestGetThresholds(t *testing.T) {
	mockMarketRepo := new(MockMarketRepository)
	mockAlertRepo := new(MockAlertRepository)

	analysisSvc := NewAnalysisService(mockMarketRepo, mockAlertRepo)

	thresholds, err := analysisSvc.GetThresholds(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, thresholds)
	assert.Contains(t, thresholds, domain.AlertTypeWashTrading)
	assert.Contains(t, thresholds, domain.AlertTypeSpoofing)
	assert.Contains(t, thresholds, domain.AlertTypePriceManipulation)
	assert.Contains(t, thresholds, domain.AlertTypeVolumeAnomaly)
}

// Helper to get string representation of AlertType
func (a AlertType) String() string {
	return string(a)
}
