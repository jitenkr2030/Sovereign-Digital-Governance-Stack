package sanctions

import (
	"context"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/csic/transaction-monitoring/internal/config"
	"github.com/csic/transaction-monitoring/internal/domain/models"
	"github.com/csic/transaction-monitoring/internal/repository"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// SanctionsService handles sanctions screening operations
type SanctionsService struct {
	cfg            *config.Config
	repo           *repository.Repository
	cache          *repository.CacheRepository
	logger         *zap.Logger
	stopChan       chan struct{}
	wg             sync.WaitGroup
	mu             sync.RWMutex

	// Sanctions data
	sdnList        map[string]*SDNEntry      // OFAC SDN list
	sdnAddresses   map[string][]string       // Address -> SDN IDs
	euList         map[string]*EUEntry       // EU sanctions list
	unList         map[string]*UNEntry       // UN sanctions list
	bloomFilter    *BloomFilter              // Probabilistic filter for addresses
	lastRefresh    time.Time
	isInitialized  bool
}

// SDNEntry represents an OFAC SDN entry
type SDNEntry struct {
	EntNum      int      `json:"ent_num"`
	SDNName     string   `json:"sdn_name"`
	SDNType     string   `json:"sdn_type"`
	Program     string   `json:"program"`
	Title       string   `json:"title"`
	CallSign    string   `json:"call_sign"`
	VessType    string   `json:"vess_type"`
	Tonnage     string   `json:"tonnage"`
	GRT         string   `json:"grt"`
	VessFlag    string   `json:"vess_flag"`
	VessOwner   string   `json:"vess_owner"`
	Remarks     string   `json:"remarks"`
	Addresses   []string `json:"addresses"`
	Aliases     []string `json:"aliases"`
}

// EUEntry represents an EU sanctions list entry
type EUEntry struct {
	EntityID     string   `json:"entity_id"`
	Name         string   `json:"name"`
	Reason       string   `json:"reason"`
	Designation  string   `json:"designation"`
	Program      string   `json:"program"`
	Addresses    []string `json:"addresses"`
}

// UNEntry represents a UN sanctions list entry
type UNEntry struct {
	EntityID     string   `json:"entity_id"`
	Name         string   `json:"name"`
	Designation  string   `json:"designation"`
	Addresses    []string `json:"addresses"`
}

// ScreeningResult represents the result of a sanctions screening
type ScreeningResult struct {
	Address       string                `json:"address"`
	Network       models.Network        `json:"network"`
	IsSanctioned  bool                  `json:"is_sanctioned"`
	Matches       []SanctionsMatch      `json:"matches,omitempty"`
	DirectLinks   int                   `json:"direct_links"`
	IndirectLinks int                   `json:"indirect_links"`
	ScreenedAt    time.Time             `json:"screened_at"`
}

// SanctionsMatch represents a match against a sanctions list
type SanctionsMatch struct {
	List        string  `json:"list"`        // OFAC, EU, UN
	EntityID    string  `json:"entity_id"`
	EntityName  string  `json:"entity_name"`
	MatchType   string  `json:"match_type"`  // direct, indirect, owner
	MatchScore  float64 `json:"match_score"`
	Program     string  `json:"program"`
	Address     string  `json:"address"`
}

// BloomFilter is a simple bloom filter for address screening
type BloomFilter struct {
	bitmap    []byte
	k         int
	m         int
	numHashes int
}

// NewSanctionsService creates a new sanctions service
func NewSanctionsService(
	cfg *config.Config,
	repo *repository.Repository,
	cache *repository.CacheRepository,
	logger *zap.Logger,
) *SanctionsService {
	return &SanctionsService{
		cfg:          cfg,
		repo:         repo,
		cache:        cache,
		logger:       logger,
		stopChan:     make(chan struct{}),
		sdnList:      make(map[string]*SDNEntry),
		sdnAddresses: make(map[string][]string),
		euList:       make(map[string]*EUEntry),
		unList:       make(map[string]*UNEntry),
	}
}

// Start initializes and starts the sanctions service
func (s *SanctionsService) Start(ctx context.Context) error {
	s.logger.Info("Starting sanctions service")

	// Initialize bloom filter
	s.bloomFilter = NewBloomFilter(1000000, 0.01)

	// Load initial sanctions data
	if err := s.RefreshSanctionsLists(ctx); err != nil {
		s.logger.Error("Failed to load initial sanctions data", zap.Error(err))
	}

	// Start background refresh
	s.wg.Add(1)
	go s.refreshLoop(ctx)

	s.isInitialized = true
	return nil
}

// Stop gracefully stops the sanctions service
func (s *SanctionsService) Stop() {
	s.logger.Info("Stopping sanctions service")
	close(s.stopChan)
	s.wg.Wait()
	s.isInitialized = false
}

// refreshLoop periodically refreshes sanctions lists
func (s *SanctionsService) refreshLoop(ctx context.Context) {
	defer s.wg.Done()

	refreshInterval, _ := time.ParseDuration(s.cfg.RiskScoring.RefreshInterval)
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.RefreshSanctionsLists(ctx); err != nil {
				s.logger.Error("Failed to refresh sanctions lists", zap.Error(err))
			}
		}
	}
}

// RefreshSanctionsLists downloads and parses sanctions lists
func (s *SanctionsService) RefreshSanctionsLists(ctx context.Context) error {
	s.logger.Info("Refreshing sanctions lists")

	refreshStart := time.Now()

	// Fetch OFAC SDN list
	if err := s.fetchOFACSDNList(ctx); err != nil {
		s.logger.Warn("Failed to fetch OFAC SDN list", zap.Error(err))
	}

	// Fetch EU sanctions list
	if err := s.fetchEUSanctionsList(ctx); err != nil {
		s.logger.Warn("Failed to fetch EU sanctions list", zap.Error(err))
	}

	// Fetch UN sanctions list
	if err := s.fetchUNSanctionsList(ctx); err != nil {
		s.logger.Warn("Failed to fetch UN sanctions list", zap.Error(err))
	}

	// Update bloom filter
	s.updateBloomFilter()

	s.lastRefresh = time.Now()
	s.logger.Info("Sanctions lists refreshed",
		zap.Duration("duration", time.Since(refreshStart)),
		zap.Int("sdn_entries", len(s.sdnList)),
		zap.Int("eu_entries", len(s.euList)),
		zap.Int("un_entries", len(s.unList)))

	return nil
}

// fetchOFACSDNList downloads and parses the OFAC SDN list
func (s *SanctionsService) fetchOFACSDNList(ctx context.Context) error {
	url := "https://www.treasury.gov/ofac/downloads/add.csv"

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch OFAC list: %w", err)
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp)
	reader.Comma = ','

	// Skip header
	_, err = reader.Read()
	if err != nil {
		return fmt.Errorf("failed to parse OFAC CSV header: %w", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		// Parse SDN entry (simplified)
		entry := &SDNEntry{
			EntNum:   parseIntOrZero(record[0]),
			SDNName:  record[1],
			SDNType:  record[2],
			Program:  record[3],
			Title:    record[4],
			CallSign: record[5],
			VessType: record[6],
			Tonnage:  record[7],
			GRT:      record[8],
			VessFlag: record[9],
			VessOwner: record[10],
			Remarks:  record[11],
		}

		if entry.EntNum > 0 {
			s.sdnList[fmt.Sprintf("%d", entry.EntNum)] = entry

			// Add addresses to mapping
			if len(record) > 12 {
				addresses := strings.Split(record[12], ";")
				for _, addr := range addresses {
					addr = strings.TrimSpace(addr)
					if addr != "" {
						s.sdnAddresses[strings.ToLower(addr)] = append(
							s.sdnAddresses[strings.ToLower(addr)],
							fmt.Sprintf("%d", entry.EntNum),
						)
					}
				}
			}
		}
	}

	return nil
}

// fetchEUSanctionsList fetches EU sanctions list (placeholder)
func (s *SanctionsService) fetchEUSanctionsList(ctx context.Context) error {
	// Implementation would fetch from EU sanctions list source
	// For now, this is a placeholder
	s.logger.Info("EU sanctions list fetch (placeholder)")
	return nil
}

// fetchUNSanctionsList fetches UN sanctions list (placeholder)
func (s *SanctionsService) fetchUNSanctionsList(ctx context.Context) error {
	// Implementation would fetch from UN consolidated list
	// For now, this is a placeholder
	s.logger.Info("UN sanctions list fetch (placeholder)")
	return nil
}

// updateBloomFilter updates the bloom filter with current sanctions addresses
func (s *SanctionsService) updateBloomFilter() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for addr := range s.sdnAddresses {
		s.bloomFilter.Add(addr)
	}
}

// ScreenAddress screens a single address against all sanctions lists
func (s *SanctionsService) ScreenAddress(ctx context.Context, address string, network models.Network) (*ScreeningResult, error) {
	result := &ScreeningResult{
		Address:     address,
		Network:     network,
		IsSanctioned: false,
		Matches:     []SanctionsMatch{},
		ScreenedAt:  time.Now(),
	}

	// Normalize address based on network
	normalizedAddr := s.normalizeAddress(address, network)
	if normalizedAddr == "" {
		return result, fmt.Errorf("invalid address format")
	}

	// Check bloom filter first (fast negative check)
	if !s.bloomFilter.Test(normalizedAddr) {
		// Likely not sanctioned
		return result, nil
	}

	// Check OFAC SDN list
	if matches, ok := s.sdnAddresses[normalizedAddr]; ok {
		for _, entNum := range matches {
			if entry, exists := s.sdnList[entNum]; exists {
				result.IsSanctioned = true
				result.Matches = append(result.Matches, SanctionsMatch{
					List:       "OFAC",
					EntityID:   entNum,
					EntityName: entry.SDNName,
					MatchType:  "direct",
					MatchScore: 1.0,
					Program:    entry.Program,
					Address:    normalizedAddr,
				})
			}
		}
	}

	// Check for indirect links (transactions with sanctioned entities)
	directLinks, indirectLinks, err := s.checkTransactionLinks(ctx, address, network)
	if err != nil {
		s.logger.Warn("Failed to check transaction links", zap.Error(err))
	}
	result.DirectLinks = directLinks
	result.IndirectLinks = indirectLinks

	if indirectLinks > 0 {
		result.Matches = append(result.Matches, SanctionsMatch{
			List:       "Indirect",
			EntityID:   "",
			EntityName: "Linked to sanctioned entity",
			MatchType:  "indirect",
			MatchScore: float64(indirectLinks) / float64(indirectLinks+1),
		})
	}

	return result, nil
}

// ScreenTransaction screens all addresses in a transaction
func (s *SanctionsService) ScreenTransaction(ctx context.Context, tx *models.NormalizedTransaction) ([]*ScreeningResult, error) {
	var results []*ScreeningResult

	for _, input := range tx.Inputs {
		result, err := s.ScreenAddress(ctx, input.Address, tx.Network)
		if err != nil {
			s.logger.Warn("Failed to screen input address",
				zap.String("address", input.Address),
				zap.Error(err))
			continue
		}
		results = append(results, result)
	}

	for _, output := range tx.Outputs {
		result, err := s.ScreenAddress(ctx, output.Address, tx.Network)
		if err != nil {
			s.logger.Warn("Failed to screen output address",
				zap.String("address", output.Address),
				zap.Error(err))
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// checkTransactionLinks checks if an address has transacted with sanctioned entities
func (s *SanctionsService) checkTransactionLinks(ctx context.Context, address string, network models.Network) (direct, indirect int, err error) {
	// Get transaction history
	txs, err := s.repo.GetTransactionHistory(ctx, address, network, 100)
	if err != nil {
		return 0, 0, err
	}

	// Check counterparties
	counterparties := make(map[string]bool)
	for _, tx := range txs {
		if tx.Sender == address {
			counterparties[tx.Receiver] = true
		} else {
			counterparties[tx.Sender] = true
		}
	}

	// Check each counterparty against sanctions
	for cp := range counterparties {
		normalized := s.normalizeAddress(cp, network)
		if normalized == "" {
			continue
		}

		if _, ok := s.sdnAddresses[normalized]; ok {
			direct++
		} else {
			// Check if counterparty is linked to sanctioned entity
			isLinked, err := s.isLinkedToSanctioned(ctx, cp, network)
			if err == nil && isLinked {
				indirect++
			}
		}
	}

	return direct, indirect, nil
}

// isLinkedToSanctioned checks if an address is linked to a sanctioned entity
func (s *SanctionsService) isLinkedToSanctioned(ctx context.Context, address string, network models.Network) (bool, error) {
	// Check via clustering
	cluster, err := s.repo.GetWalletByAddress(ctx, address, network)
	if err != nil {
		return false, err
	}

	if cluster.ClusterID != nil {
		c, err := s.repo.GetClusterByID(ctx, *cluster.ClusterID)
		if err != nil {
			return false, err
		}

		members, err := s.repo.GetClusterMembers(ctx, *cluster.ClusterID)
		if err != nil {
			return false, err
		}

		for _, member := range members {
			normalized := s.normalizeAddress(member.WalletAddress, member.Network)
			if normalized != "" && s.bloomFilter.Test(normalized) {
				return true, nil
			}
		}
	}

	return false, nil
}

// normalizeAddress normalizes an address based on network
func (s *SanctionsService) normalizeAddress(address string, network models.Network) string {
	switch network {
	case models.NetworkEthereum:
		if !strings.HasPrefix(address, "0x") {
			address = "0x" + address
		}
		if !common.IsHexAddress(address) {
			return ""
		}
		return strings.ToLower(address)

	case models.NetworkBitcoin:
		// Basic Bitcoin address normalization
		return strings.ToLower(address)

	default:
		return strings.ToLower(address)
	}
}

// CreateSanctionsAlert creates an alert for sanctions match
func (s *SanctionsService) CreateSanctionsAlert(ctx context.Context, result *ScreeningResult) error {
	if !result.IsSanctioned && result.DirectLinks == 0 {
		return nil
	}

	alert := models.NewAlert(
		models.AlertTypeSanctionsMatch,
		models.SeverityCritical,
		"wallet",
		result.Address,
		fmt.Sprintf("Sanctions match detected: %s", result.Address),
	)

	alert.RuleName = "sanctions_screening"
	alert.Score = 100
	if len(result.Matches) > 0 {
		alert.Evidence = []string{result.Matches[0].EntityName}
	}

	// Set metadata
	metadata := map[string]interface{}{
		"network":        result.Network,
		"direct_links":   result.DirectLinks,
		"indirect_links": result.IndirectLinks,
		"matches":        result.Matches,
	}
	alert.Metadata = metadata

	return s.repo.CreateAlert(ctx, alert)
}

// GetSanctionsStats returns sanctions screening statistics
func (s *SanctionsService) GetSanctionsStats(ctx context.Context) (*SanctionsStats, error) {
	stats := &SanctionsStats{
		LastRefresh: s.lastRefresh,
		TotalSDN:    len(s.sdnList),
		TotalEU:     len(s.euList),
		TotalUN:     len(s.unList),
	}

	// Get alert statistics from database
	alerts, err := s.repo.GetAlertsByType(ctx, models.AlertTypeSanctionsMatch)
	if err != nil {
		return nil, err
	}

	stats.PendingAlerts = len(alerts)

	return stats, nil
}

// SanctionsStats contains sanctions service statistics
type SanctionsStats struct {
	LastRefresh    time.Time `json:"last_refresh"`
	TotalSDN       int       `json:"total_sdn_entries"`
	TotalEU        int       `json:"total_eu_entries"`
	TotalUN        int       `json:"total_un_entries"`
	PendingAlerts  int       `json:"pending_alerts"`
}

// HealthCheck verifies the service is healthy
func (s *SanctionsService) HealthCheck(ctx context.Context) error {
	if !s.isInitialized {
		return fmt.Errorf("sanctions service not initialized")
	}

	// Check if lists are fresh
	maxAge, _ := time.ParseDuration(s.cfg.RiskScoring.RefreshInterval)
	if time.Since(s.lastRefresh) > maxAge*2 {
		return fmt.Errorf("sanctions lists are stale")
	}

	return nil
}

// NewBloomFilter creates a new bloom filter
func NewBloomFilter(expectedItems int, falsePositiveRate float64) *BloomFilter {
	m := calculateM(expectedItems, falsePositiveRate)
	k := calculateK(expectedItems, m)

	return &BloomFilter{
		bitmap:    make([]byte, m),
		k:         k,
		m:         m,
		numHashes: int(k),
	}
}

// Add adds an item to the bloom filter
func (bf *BloomFilter) Add(item string) {
	for i := 0; i < bf.numHashes; i++ {
		index := bf.hash(item, i) % bf.m
		bf.bitmap[index/8] |= 1 << (index % 8)
	}
}

// Test checks if an item might be in the set
func (bf *BloomFilter) Test(item string) bool {
	for i := 0; i < bf.numHashes; i++ {
		index := bf.hash(item, i) % bf.m
		if (bf.bitmap[index/8] & (1 << (index % 8))) == 0 {
			return false
		}
	}
	return true
}

// hash calculates hash values for bloom filter
func (bf *BloomFilter) hash(item string, seed int) uint64 {
	h := []byte(item)
	return uint64(hex.EncodeToString(h)[seed%len(h)])
}

func calculateM(expectedItems int, falsePositiveRate float64) int {
	return int(-float64(expectedItems) * math.Log(falsePositiveRate) / (math.Ln2 * math.Ln2))
}

func calculateK(expectedItems, m int) float64 {
	return (float64(m) / float64(expectedItems)) * math.Ln2
}

func parseIntOrZero(s string) int {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		result = result*10 + int(c-'0')
	}
	return result
}
