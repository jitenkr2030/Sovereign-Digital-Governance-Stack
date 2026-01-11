module.exports = {
  rules: {
    'energy-adapter-down': {
      alert: 'EnergyAdapterDown',
      expr: 'up{job="energy-adapter"} == 0',
      for: '5m',
      labels: {
        severity: 'critical',
        team: 'neam-platform'
      },
      annotations: {
        summary: 'Energy adapter is down',
        description: 'The energy adapter instance {{ $labels.instance }} has been down for more than 5 minutes'
      }
    },
    'energy-adapter-high-latency': {
      alert: 'EnergyAdapterHighLatency',
      expr: 'histogram_quantile(0.95, rate(energy_adapter_processing_latency_seconds_bucket[5m])) > 0.01',
      for: '5m',
      labels: {
        severity: 'warning',
        team: 'neam-platform'
      },
      annotations: {
        summary: 'Energy adapter high latency',
        description: 'Processing latency is above 10ms for {{ $labels.instance }}'
      }
    },
    'energy-adapter-high-filter-rate': {
      alert: 'EnergyAdapterHighFilterRate',
      expr: 'rate(energy_adapter_points_filtered_total[5m]) / rate(energy_adapter_points_processed_total[5m]) > 0.5',
      for: '5m',
      labels: {
        severity: 'warning',
        team: 'neam-platform'
      },
      annotations: {
        summary: 'High filter rate detected',
        description: 'More than 50% of points are being filtered for {{ $labels.instance }}'
      }
    },
    'energy-adapter-connection-unhealthy': {
      alert: 'EnergyAdapterConnectionUnhealthy',
      expr: 'energy_adapter_connection_health < 1',
      for: '1m',
      labels: {
        severity: 'warning',
        team: 'neam-platform'
      },
      annotations: {
        summary: 'Unhealthy connection detected',
        description: 'Connection {{ $labels.connection_id }} is unhealthy'
      }
    },
    'energy-adapter-kafka-failure': {
      alert: 'EnergyAdapterKafkaFailure',
      expr: 'increase(kafka_producer_messages_failed_total[5m]) > 10',
      for: '2m',
      labels: {
        severity: 'critical',
        team: 'neam-platform'
      },
      annotations: {
        summary: 'Kafka message failures',
        description: 'More than 10 Kafka message failures in the last 5 minutes'
      }
    }
  }
};
