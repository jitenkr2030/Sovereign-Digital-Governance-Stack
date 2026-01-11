package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/neamplatform/sensing/energy/config"
	"github.com/neamplatform/sensing/energy/kafka"
	"github.com/neamplatform/sensing/energy/metrics"
	"github.com/neamplatform/sensing/energy/signalproc"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

func init() {
	// Initialize custom metrics
	_ = promauto.NewCounter(prometheus.CounterOpts{
		Name: "energy_adapter_initialized_total",
		Help: "Counter for adapter initialization",
	})
}

func main() {
	// This is a placeholder for metrics documentation
	// The actual metrics are defined in metrics/collector.go

	fmt.Println("Energy Adapter Metrics")
	fmt.Println("======================")
	fmt.Println("")
	fmt.Println("Available metrics:")
	fmt.Println("")
	fmt.Println("  - energy_adapter_points_processed_total")
	fmt.Println("  - energy_adapter_points_filtered_total")
	fmt.Println("  - energy_adapter_spikes_detected_total")
	fmt.Println("  - energy_adapter_anomalies_detected_total")
	fmt.Println("  - energy_adapter_quality_filtered_total")
	fmt.Println("  - energy_adapter_processing_latency_seconds")
	fmt.Println("  - energy_adapter_connection_health")
	fmt.Println("  - kafka_producer_messages_sent_total")
	fmt.Println("  - kafka_producer_messages_failed_total")
	fmt.Println("  - kafka_producer_bytes_sent_total")
	fmt.Println("  - kafka_producer_write_latency_seconds")
	fmt.Println("  - energy_adapter_active_connections")
	fmt.Println("  - energy_adapter_session_retries_total")
	fmt.Println("  - energy_adapter_session_state")
	fmt.Println("  - energy_adapter_pool_size")
	fmt.Println("  - energy_adapter_pool_capacity")
	fmt.Println("")
}

// Example usage functions for documentation purposes

func exampleUsage() {
	// Example: Recording metrics
	logger, _ := zap.NewDevelopment()

	// Start metrics server
	metricsServer := metrics.DefaultServer(logger)
	metricsServer.Start()

	// Record a processed point
	metrics.RecordPointProcessed("source-1", "voltage-1", "ICCP")

	// Record a filtered point
	metrics.RecordPointFiltered("source-1", "voltage-1", "ICCP", "deadband")

	// Record processing latency
	metrics.RecordProcessingLatency("filter", 0.001)

	// Set connection health
	metrics.SetConnectionHealth("connection-1", "ICCP", true)

	// Record Kafka metrics
	metrics.RecordKafkaMessageSent()
	metrics.RecordKafkaBytesSent(1024)
	metrics.RecordKafkaLatency(0.005)

	// Set active connections
	metrics.SetActiveConnections("ICCP", 2.0)

	// Stop metrics server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	metricsServer.Stop(ctx)
}

func exampleConfig() {
	// Example configuration
	cfg := config.Default()
	_ = cfg

	// Kafka producer example
	kafkaCfg := config.KafkaConfig{
		Brokers:      "localhost:9092",
		OutputTopic:  "energy.telemetry",
		DLQTopic:     "energy.dlq",
		Version:      "2.8.0",
		MaxOpenReqs:  100,
		DialTimeout:  config.Duration(10 * time.Second),
		WriteTimeout: config.Duration(30 * time.Second),
	}

	logger, _ := zap.NewDevelopment()

	producer, err := kafka.NewProducer(kafkaCfg, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create producer: %v\n", err)
		return
	}
	defer producer.Close()

	// Produce a message
	ctx := context.Background()
	err = producer.Produce(ctx, kafkaCfg.OutputTopic, map[string]interface{}{
		"source": "example",
		"value":  42.0,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to produce: %v\n", err)
		return
	}

	_ = producer
}

func exampleSignalProcessing() {
	// Example signal processing configuration
	cfg := config.FilteringConfig{
		DeadbandPercent:     0.1,
		SpikeThreshold:      20.0,
		SpikeWindowSize:     5,
		RateOfChangeLimit:   5.0,
		QualityFilter:       []string{"Good"},
		MinReportingInterval: config.Duration(100 * time.Millisecond),
		MaxValue:            1000000.0,
		MinValue:            -1000000.0,
	}

	logger, _ := zap.NewDevelopment()

	processor := signalproc.NewProcessor(cfg, logger)

	// Process telemetry
	input := make(chan signalproc.TelemetryPoint, 100)
	output := processor.Process(input)

	// Send points
	go func() {
		for i := 0; i < 10; i++ {
			input <- signalproc.TelemetryPoint{
				ID:        fmt.Sprintf("point-%d", i),
				SourceID:  "example-source",
				PointID:   "voltage",
				Value:     230.0 + float64(i),
				Quality:   "Good",
				Protocol:  "ICCP",
				Timestamp: time.Now(),
			}
		}
		close(input)
	}()

	// Receive processed points
	for point := range output {
		fmt.Printf("Processed point: %s = %f\n", point.PointID, point.Value)
	}

	// Get statistics
	stats := processor.Stats()
	fmt.Printf("Processed: %d, Filtered: %d\n", stats.PointsProcessed, stats.PointsFiltered)
}

// Import prometheus for init function is already included above
