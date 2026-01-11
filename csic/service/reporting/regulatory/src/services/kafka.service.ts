/**
 * Kafka Service
 * 
 * Handles Kafka producer operations for event publishing.
 */

import { Kafka, Producer, logLevel } from 'kafkajs';
import { config } from '../config/config';

export interface KafkaEvent {
  eventType: string;
  payload: any;
  timestamp: string;
  source: string;
}

export class KafkaService {
  private kafka: Kafka;
  private producer: Producer;
  private isConnected: boolean = false;

  constructor() {
    this.kafka = new Kafka({
      clientId: config.kafka.client_id,
      brokers: config.kafka.brokers,
      logLevel: logLevel.WARN,
      retry: {
        initialRetryTime: 100,
        retries: 8,
      },
    });

    this.producer = this.kafka.producer();
  }

  async connect(): Promise<void> {
    try {
      await this.producer.connect();
      this.isConnected = true;
      console.log('Kafka producer connected');
    } catch (error) {
      console.error('Failed to connect to Kafka:', error);
      // Don't throw - Kafka is optional for some operations
    }
  }

  async disconnect(): Promise<void> {
    try {
      await this.producer.disconnect();
      this.isConnected = false;
      console.log('Kafka producer disconnected');
    } catch (error) {
      console.error('Error disconnecting from Kafka:', error);
    }
  }

  async publishEvent(eventType: string, payload: any): Promise<void> {
    if (!this.isConnected) {
      console.warn('Kafka not connected, skipping event publish');
      return;
    }

    const event: KafkaEvent = {
      eventType,
      payload,
      timestamp: new Date().toISOString(),
      source: 'regulatory-reporting-service',
    };

    const topic = this.getTopicForEvent(eventType);

    try {
      await this.producer.send({
        topic,
        messages: [
          {
            key: payload.reportId || payload.scheduleId || payload.entityId,
            value: JSON.stringify(event),
            headers: {
              eventType,
              timestamp: Date.now().toString(),
            },
          },
        ],
      });

      console.log(`Published event ${eventType} to topic ${topic}`);
    } catch (error) {
      console.error(`Failed to publish event ${eventType}:`, error);
    }
  }

  async publishBatch(events: KafkaEvent[]): Promise<void> {
    if (!this.isConnected) {
      console.warn('Kafka not connected, skipping batch event publish');
      return;
    }

    const messages = events.map((event) => ({
      topic: this.getTopicForEvent(event.eventType),
      messages: [
        {
          key: event.payload.reportId || event.payload.scheduleId || event.payload.entityId,
          value: JSON.stringify(event),
          headers: {
            eventType: event.eventType,
            timestamp: Date.now().toString(),
          },
        },
      ],
    }));

    try {
      await this.producer.sendBatch({ topics: messages });
      console.log(`Published batch of ${events.length} events`);
    } catch (error) {
      console.error('Failed to publish batch events:', error);
    }
  }

  private getTopicForEvent(eventType: string): string {
    const topicMap: Record<string, string> = {
      'report.generated': config.kafka.topics.report_generated,
      'report.failed': config.kafka.topics.report_failed,
      'report.scheduled': 'report.scheduled',
      'report.downloaded': 'report.downloaded',
    };

    return topicMap[eventType] || 'report.events';
  }

  isHealthy(): boolean {
    return this.isConnected;
  }
}

// Export singleton instance
export const kafkaService = new KafkaService();
