import { Injectable, OnModuleInit, OnModuleDestroy, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { Kafka, Producer, AdminClient, logLevel } from 'kafkajs';

@Injectable()
export class KafkaService implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(KafkaService.name);
  private kafka: Kafka;
  private producer: Producer;
  private admin: AdminClient;
  private isConnected: boolean = false;

  constructor(private configService: ConfigService) {
    const clientId = this.configService.get('KAFKA_CLIENT_ID', 'licensing-service');
    const brokers = this.configService.get('KAFKA_BROKERS', 'localhost:9092').split(',');

    this.kafka = new Kafka({
      clientId,
      brokers,
      logLevel: logLevel.WARN,
      retry: {
        initialRetryTime: 100,
        retries: 8,
      },
    });

    this.producer = this.kafka.producer();
    this.admin = this.kafka.admin();
  }

  async onModuleInit() {
    try {
      await this.admin.connect();
      await this.ensureTopicsExist();
      await this.producer.connect();
      this.isConnected = true;
      this.logger.log('Kafka producer connected successfully');
    } catch (error) {
      this.logger.error('Failed to connect to Kafka', error);
      // Don't block application startup - Kafka is optional
    }
  }

  async onModuleDestroy() {
    try {
      await this.producer.disconnect();
      await this.admin.disconnect();
      this.isConnected = false;
      this.logger.log('Kafka producer disconnected');
    } catch (error) {
      this.logger.error('Error disconnecting from Kafka', error);
    }
  }

  private async ensureTopicsExist() {
    const topicPrefix = this.configService.get('KAFKA_TOPIC_PREFIX', 'license');
    const topics = [
      `${topicPrefix}.application.submitted`,
      `${topicPrefix}.application.approved`,
      `${topicPrefix}.application.rejected`,
      `${topicPrefix}.license.issued`,
      `${topicPrefix}.license.expiring`,
      `${topicPrefix}.license.expired`,
      `${topicPrefix}.license.suspended`,
      `${topicPrefix}.license.revoked`,
    ];

    try {
      const existingTopics = await this.admin.listTopics();
      const topicsToCreate = topics.filter(t => !existingTopics.includes(t));

      if (topicsToCreate.length > 0) {
        await this.admin.createTopics({
          topics: topicsToCreate.map(topic => ({
            topic,
            numPartitions: 3,
            replicationFactor: 1,
          })),
        });
        this.logger.log(`Created topics: ${topicsToCreate.join(', ')}`);
      }
    } catch (error) {
      this.logger.warn('Error creating Kafka topics (may already exist)', error);
    }
  }

  async publishLicenseEvent(eventType: string, payload: any): Promise<void> {
    if (!this.isConnected) {
      this.logger.warn('Kafka not connected, skipping event publish');
      return;
    }

    const topicPrefix = this.configService.get('KAFKA_TOPIC_PREFIX', 'license');
    const topic = eventType.startsWith('license.')
      ? eventType
      : `${topicPrefix}.${eventType}`;

    const message = {
      key: payload.applicationId || payload.licenseId || payload.entityId,
      value: JSON.stringify({
        eventType,
        payload,
        timestamp: new Date().toISOString(),
        source: 'licensing-service',
      }),
      headers: {
        eventType,
        timestamp: Date.now().toString(),
      },
    };

    try {
      await this.producer.send({
        topic,
        messages: [message],
      });
      this.logger.log(`Published event ${eventType} to topic ${topic}`);
    } catch (error) {
      this.logger.error(`Failed to publish event ${eventType}`, error);
      throw error;
    }
  }

  async publishBatch(events: Array<{ eventType: string; payload: any }>): Promise<void> {
    if (!this.isConnected) {
      this.logger.warn('Kafka not connected, skipping batch event publish');
      return;
    }

    const topicPrefix = this.configService.get('KAFKA_TOPIC_PREFIX', 'license');

    const messages = events.map(event => {
      const topic = event.eventType.startsWith('license.')
        ? event.eventType
        : `${topicPrefix}.${event.eventType}`;

      return {
        topic,
        key: event.payload.applicationId || event.payload.licenseId || event.payload.entityId,
        value: JSON.stringify({
          eventType: event.eventType,
          payload: event.payload,
          timestamp: new Date().toISOString(),
          source: 'licensing-service',
        }),
        headers: {
          eventType: event.eventType,
          timestamp: Date.now().toString(),
        },
      };
    });

    try {
      await this.producer.send({ messages });
      this.logger.log(`Published batch of ${events.length} events`);
    } catch (error) {
      this.logger.error('Failed to publish batch events', error);
      throw error;
    }
  }

  isHealthy(): boolean {
    return this.isConnected;
  }
}
