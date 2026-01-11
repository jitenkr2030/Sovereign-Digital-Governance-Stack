package coordination

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"neam-platform/shared"
)

// TestMetricsCollector tests the metrics collector
func TestMetricsCollector(t *testing.T) {
	logger := shared.NewLogger(nil)
	collector := NewMetricsCollector(logger)

	t.Run("new collector has zero metrics", func(t *testing.T) {
		metrics := collector.GetGlobalMetrics()

		assert.Equal(t, int64(0), metrics.TotalAcquisitions)
		assert.Equal(t, int64(0), metrics.TotalReleases)
		assert.Equal(t, int64(0), metrics.TotalContentions)
		assert.Equal(t, int64(0), metrics.TotalDeadlocks)
	})

	t.Run("record acquisition success", func(t *testing.T) {
		collector.RecordAcquisition("test", 10*time.Millisecond, true)

		metrics := collector.GetGlobalMetrics()
		assert.Equal(t, int64(1), metrics.TotalAcquisitions)
		assert.True(t, metrics.AvgAcquisitionTime > 0)
	})

	t.Run("record acquisition failure", func(t *testing.T) {
		collector.RecordAcquisition("test", 50*time.Millisecond, false)

		metrics := collector.GetGlobalMetrics()
		assert.Equal(t, int64(2), metrics.TotalAcquisitions)
	})

	t.Run("record release", func(t *testing.T) {
		collector.RecordRelease("test", 100*time.Millisecond)

		metrics := collector.GetGlobalMetrics()
		assert.Equal(t, int64(1), metrics.TotalReleases)
		assert.True(t, metrics.AvgHoldTime > 0)
	})

	t.Run("record contention", func(t *testing.T) {
		collector.RecordContention("test")

		metrics := collector.GetGlobalMetrics()
		assert.Equal(t, int64(1), metrics.TotalContentions)
	})

	t.Run("record deadlock", func(t *testing.T) {
		collector.RecordDeadlock("test")

		metrics := collector.GetGlobalMetrics()
		assert.Equal(t, int64(1), metrics.TotalDeadlocks)
	})

	t.Run("record retry", func(t *testing.T) {
		collector.RecordRetry("test", 1)
		collector.RecordRetry("test", 2)
		collector.RecordRetry("test", 3)

		resourceMetrics := collector.GetResourceMetrics("test")
		assert.Equal(t, int64(3), resourceMetrics.retryAttempts)
	})
}

// TestResourceMetrics tests per-resource metrics
func TestResourceMetrics(t *testing.T) {
	logger := shared.NewLogger(nil)
	collector := NewMetricsCollector(logger)

	t.Run("get resource metrics", func(t *testing.T) {
		metrics := collector.GetResourceMetrics("test")

		assert.Equal(t, "test", metrics.ResourceType)
		assert.Equal(t, int64(0), metrics.acquisitions)
	})

	t.Run("multiple resource types", func(t *testing.T) {
		collector.RecordAcquisition("type1", 10*time.Millisecond, true)
		collector.RecordAcquisition("type2", 20*time.Millisecond, true)
		collector.RecordAcquisition("type1", 30*time.Millisecond, true)

		metrics1 := collector.GetResourceMetrics("type1")
		metrics2 := collector.GetResourceMetrics("type2")

		assert.Equal(t, int64(2), metrics1.acquisitions)
		assert.Equal(t, int64(1), metrics2.acquisitions)
	})

	t.Run("get all resource metrics", func(t *testing.T) {
		collector.RecordAcquisition("type1", 10*time.Millisecond, true)
		collector.RecordAcquisition("type2", 20*time.Millisecond, true)

		allMetrics := collector.GetAllResourceMetrics()
		assert.Len(t, allMetrics, 2)
	})
}

// TestContentiousResource tests contentious resource tracking
func TestContentiousResource(t *testing.T) {
	logger := shared.NewLogger(nil)
	collector := NewMetricsCollector(logger)

	t.Run("record first contention", func(t *testing.T) {
		collector.RecordContentiousResource("test", "resource1", 50*time.Millisecond)

		topResources := collector.GetTopContentiousResources("test", 10)
		assert.Len(t, topResources, 1)
		assert.Equal(t, "resource1", topResources[0].ResourceID)
		assert.Equal(t, int64(1), topResources[0].ContentionCount)
	})

	t.Run("record multiple contentions", func(t *testing.T) {
		collector.RecordContentiousResource("test", "resource1", 50*time.Millisecond)
		collector.RecordContentiousResource("test", "resource1", 100*time.Millisecond)
		collector.RecordContentiousResource("test", "resource2", 200*time.Millisecond)

		topResources := collector.GetTopContentiousResources("test", 10)
		assert.Len(t, topResources, 2)

		// resource1 should be first (more contentions)
		assert.Equal(t, "resource1", topResources[0].ResourceID)
		assert.Equal(t, int64(2), topResources[0].ContentionCount)
	})

	t.Run("top N resources", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			collector.RecordContentiousResource("test", "resource", 50*time.Millisecond)
		}

		topResources := collector.GetTopContentiousResources("test", 5)
		assert.Len(t, topResources, 5)
	})
}

// TestContentionScore tests the contention score calculation
func TestContentionScore(t *testing.T) {
	logger := shared.NewLogger(nil)
	collector := NewMetricsCollector(logger)

	t.Run("zero score with no acquisitions", func(t *testing.T) {
		score := collector.GetContentionScore("test")
		assert.Equal(t, float64(0), score)
	})

	t.Run("low contention score", func(t *testing.T) {
		// Many acquisitions, few contentions
		for i := 0; i < 100; i++ {
			collector.RecordAcquisition("test", 10*time.Millisecond, true)
		}
		collector.RecordContention("test")

		score := collector.GetContentionScore("test")
		assert.True(t, score > 0)
		assert.True(t, score < 1)
	})

	t.Run("high contention score", func(t *testing.T) {
		// Few acquisitions, many contentions
		collector.RecordAcquisition("test", 10*time.Millisecond, true)
		collector.RecordContention("test")
		collector.RecordContention("test")
		collector.RecordContention("test")
		collector.RecordContention("test")

		score := collector.GetContentionScore("test")
		assert.True(t, score > 0.5)
	})
}

// TestMetricsReset tests metrics reset
func TestMetricsReset(t *testing.T) {
	logger := shared.NewLogger(nil)
	collector := NewMetricsCollector(logger)

	// Add some metrics
	collector.RecordAcquisition("test", 10*time.Millisecond, true)
	collector.RecordContention("test")
	collector.RecordDeadlock("test")

	// Reset
	collector.Reset()

	// Verify zeroed
	metrics := collector.GetGlobalMetrics()
	assert.Equal(t, int64(0), metrics.TotalAcquisitions)
	assert.Equal(t, int64(0), metrics.TotalContentions)
	assert.Equal(t, int64(0), metrics.TotalDeadlocks)
}

// TestMetricsSummary tests metrics summary generation
func TestMetricsSummary(t *testing.T) {
	logger := shared.NewLogger(nil)
	collector := NewMetricsCollector(logger)

	t.Run("empty summary", func(t *testing.T) {
		summaries := collector.GetSummary()
		assert.Empty(t, summaries)
	})

	t.Run("with metrics", func(t *testing.T) {
		collector.RecordAcquisition("test", 10*time.Millisecond, true)
		collector.RecordAcquisition("test", 20*time.Millisecond, true)
		collector.RecordRelease("test", 100*time.Millisecond)
		collector.RecordRelease("test", 200*time.Millisecond)
		collector.RecordContention("test")
		collector.RecordContentiousResource("test", "resource1", 50*time.Millisecond)

		summaries := collector.GetSummary()
		assert.Len(t, summaries, 1)

		summary := summaries[0]
		assert.Equal(t, "test", summary.ResourceType)
		assert.Equal(t, int64(2), summary.TotalLocks)
		assert.Equal(t, int64(1), summary.Contentions)
		assert.True(t, summary.AvgHoldTime > 0)
		assert.True(t, summary.MaxHoldTime > 0)
		assert.True(t, summary.ContentionScore >= 0)
	})
}

// TestContentionAnalyzer tests the contention analyzer
func TestContentionAnalyzer(t *testing.T) {
	logger := shared.NewLogger(nil)
	collector := NewMetricsCollector(logger)
	analyzer := NewContentionAnalyzer(collector, 0.1, 1*time.Minute)

	t.Run("no alerts with low contention", func(t *testing.T) {
		// Low contention
		for i := 0; i < 1000; i++ {
			collector.RecordAcquisition("test", 10*time.Millisecond, true)
		}
		collector.RecordContention("test")

		alerts := analyzer.AnalyzeContentiousResources()
		assert.Empty(t, alerts)
	})

	t.Run("alert with high contention", func(t *testing.T) {
		collector.Reset()

		// High contention scenario
		for i := 0; i < 10; i++ {
			collector.RecordAcquisition("test", 10*time.Millisecond, true)
		}
		collector.RecordContention("test")
		collector.RecordContention("test")
		collector.RecordContention("test")
		collector.RecordContention("test")
		collector.RecordContention("test")

		alerts := analyzer.AnalyzeContentiousResources()
		assert.NotEmpty(t, alerts)

		alert := alerts[0]
		assert.Equal(t, "test", alert.ResourceType)
		assert.Equal(t, SeverityMedium, alert.Severity)
		assert.Equal(t, int64(5), alert.ContentionCount)
	})
}

// TestAlertSeverity tests alert severity determination
func TestAlertSeverity(t *testing.T) {
	logger := shared.NewLogger(nil)
	collector := NewMetricsCollector(logger)
	analyzer := NewContentionAnalyzer(collector, 0.1, 1*time.Minute)

	t.Run("low severity", func(t *testing.T) {
		severity := analyzer.determineSeverity(0.11)
		assert.Equal(t, SeverityLow, severity)
	})

	t.Run("medium severity", func(t *testing.T) {
		severity := analyzer.determineSeverity(0.16)
		assert.Equal(t, SeverityMedium, severity)
	})

	t.Run("high severity", func(t *testing.T) {
		severity := analyzer.determineSeverity(0.21)
		assert.Equal(t, SeverityHigh, severity)
	})

	t.Run("critical severity", func(t *testing.T) {
		severity := analyzer.determineSeverity(0.31)
		assert.Equal(t, SeverityCritical, severity)
	})
}

// TestGlobalMetrics tests global metrics structure
func TestGlobalMetrics(t *testing.T) {
	metrics := GlobalMetrics{
		Timestamp:          time.Now(),
		TotalAcquisitions:  100,
		TotalReleases:      90,
		TotalContentions:   5,
		TotalDeadlocks:     1,
		AvgAcquisitionTime: 50 * time.Millisecond,
		AvgHoldTime:        200 * time.Millisecond,
	}

	assert.Equal(t, int64(100), metrics.TotalAcquisitions)
	assert.Equal(t, int64(90), metrics.TotalReleases)
	assert.Equal(t, int64(5), metrics.TotalContentions)
	assert.Equal(t, int64(1), metrics.TotalDeadlocks)
}

// TestResourceMetricsSummary tests resource metrics summary structure
func TestResourceMetricsSummary(t *testing.T) {
	summary := ResourceMetricsSummary{
		ResourceType:    "test",
		TotalLocks:      50,
		Contentions:     3,
		AvgHoldTime:     100 * time.Millisecond,
		MaxHoldTime:     500 * time.Millisecond,
		ContentionScore: 0.05,
		TopContentious:  []string{"resource1", "resource2"},
	}

	assert.Equal(t, "test", summary.ResourceType)
	assert.Equal(t, int64(50), summary.TotalLocks)
	assert.Len(t, summary.TopContentious, 2)
}

// TestContentionAlert tests contention alert structure
func TestContentionAlert(t *testing.T) {
	alert := ContentionAlert{
		ResourceType:    "test",
		ContentionScore: 0.25,
		Threshold:       0.1,
		ContentionCount: 10,
		AvgWaitTime:     100 * time.Millisecond,
		DetectedAt:      time.Now(),
		Severity:        SeverityHigh,
	}

	assert.Equal(t, "test", alert.ResourceType)
	assert.Equal(t, SeverityHigh, alert.Severity)
	assert.Equal(t, int64(10), alert.ContentionCount)
}

// TestAlertSeverityConstants tests alert severity constants
func TestAlertSeverityConstants(t *testing.T) {
	assert.Equal(t, AlertSeverity("low"), SeverityLow)
	assert.Equal(t, AlertSeverity("medium"), SeverityMedium)
	assert.Equal(t, AlertSeverity("high"), SeverityHigh)
	assert.Equal(t, AlertSeverity("critical"), SeverityCritical)
}

// TestThreadSafety tests thread safety of metrics collector
func TestThreadSafety(t *testing.T) {
	logger := shared.NewLogger(nil)
	collector := NewMetricsCollector(logger)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrently record metrics
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			collector.RecordAcquisition("test", 10*time.Millisecond, true)
			collector.RecordContention("test")
			collector.RecordRelease("test", 100*time.Millisecond)
		}(i)
	}

	wg.Wait()

	// Verify metrics were recorded
	metrics := collector.GetGlobalMetrics()
	assert.Equal(t, int64(numGoroutines), metrics.TotalAcquisitions)
	assert.Equal(t, int64(numGoroutines), metrics.TotalReleases)
	assert.Equal(t, int64(numGoroutines), metrics.TotalContentions)
}

// TestRecordExtendedHold tests extended hold recording
func TestRecordExtendedHold(t *testing.T) {
	logger := shared.NewLogger(nil)
	collector := NewMetricsCollector(logger)

	// Should not panic
	collector.RecordExtendedHold("test", 5*time.Minute, 1*time.Minute)

	// shared.Logger handles warnings internally
}
