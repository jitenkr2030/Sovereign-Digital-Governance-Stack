package main

import (
	"neam-platform/shared"
)

// ReportingConfig extends the shared configuration with reporting-specific settings
type ReportingConfig struct {
	*shared.Config
	
	// Export configuration
	ExportDir string `mapstructure:"export_dir" json:"export_dir"`
	
	// Archive configuration  
	ArchiveDir string `mapstructure:"archive_dir" json:"archive_dir"`
	
	// Forensic configuration
	EvidenceRetentionYears int `mapstructure:"evidence_retention_years" json:"evidence_retention_years"`
	
	// Real-time configuration
	MetricsRefreshInterval int `mapstructure:"metrics_refresh_interval" json:"metrics_refresh_interval"`
	AnomalyCheckInterval   int `mapstructure:"anomaly_check_interval" json:"anomaly_check_interval"`
	
	// Archive configuration
	ArchiveCompressionAlgo   string `mapstructure:"archive_compression_algo" json:"archive_compression_algo"`
	ArchiveEncryptionAlgo    string `mapstructure:"archive_encryption_algo" json:"archive_encryption_algo"`
	RetrievalPriorityDefault string `mapstructure:"retrieval_priority_default" json:"retrieval_priority_default"`
	
	// Retention policy configuration
	DailyRetentionDays   int `mapstructure:"daily_retention_days" json:"daily_retention_days"`
	MonthlyRetentionYears int `mapstructure:"monthly_retention_years" json:"monthly_retention_years"`
	QuarterlyRetentionYears int `mapstructure:"quarterly_retention_years" json:"quarterly_retention_years"`
	
	// Storage tier configuration
	HotStorageDays     int `mapstructure:"hot_storage_days" json:"hot_storage_days"`
	WarmStorageDays    int `mapstructure:"warm_storage_days" json:"warm_storage_days"`
	ColdStorageEnabled bool `mapstructure:"cold_storage_enabled" json:"cold_storage_enabled"`
}

// GetDefaultReportingConfig returns the default configuration for reporting service
func GetDefaultReportingConfig() *ReportingConfig {
	return &ReportingConfig{
		Config: shared.GetDefaultConfig(),
		
		// Export settings
		ExportDir: "/var/lib/neam/reporting/exports",
		
		// Archive settings
		ArchiveDir: "/var/lib/neam/reporting/archives",
		
		// Forensic settings
		EvidenceRetentionYears: 25,
		
		// Real-time settings
		MetricsRefreshInterval: 30,
		AnomalyCheckInterval:   60,
		
		// Archive settings
		ArchiveCompressionAlgo:  "zstd",
		ArchiveEncryptionAlgo:   "aes256",
		RetrievalPriorityDefault: "standard",
		
		// Retention settings
		DailyRetentionDays:      90,
		MonthlyRetentionYears:   10,
		QuarterlyRetentionYears: 20,
		
		// Storage tier settings
		HotStorageDays:      7,
		WarmStorageDays:     90,
		ColdStorageEnabled:  true,
	}
}
