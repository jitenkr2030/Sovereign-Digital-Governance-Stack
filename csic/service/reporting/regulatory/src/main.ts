/**
 * Regulatory Reporting Service - Main Entry Point
 * 
 * Node.js-based service for PDF/CSV report generation,
 * automated scheduling, and WORM storage compliance.
 */

import express, { Application, Request, Response, NextFunction } from 'express';
import cors from 'cors';
import helmet from 'helmet';
import morgan from 'morgan';
import { createBullBoard } from '@bull-board/express';
import Bull from 'bull';

import { config } from './config/config';
import { reportRoutes } from './routes/reports';
import { scheduleRoutes } from './routes/schedules';
import { templateRoutes } from './routes/templates';
import { healthRoutes } from './routes/health';
import { ReportQueue } from './queue/reportQueue';
import { KafkaService } from './services/kafka';
import { PrometheusService } from './services/prometheus';

// Initialize Express app
const app: Application = express();

// Initialize services
const kafkaService = new KafkaService();
const prometheusService = new PrometheusService();

// Initialize Bull queue
const reportQueue = new ReportQueue();
const bullAdapter = createBullBoard({
  queues: [new Bull.Adapter(reportQueue.getQueue())],
  serverAdapter: bullAdapter,
});

// Middleware
app.use(helmet());
app.use(cors({
  origin: config.server.corsOrigins,
  credentials: true,
}));
app.use(morgan('combined'));
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true }));

// Prometheus metrics endpoint
app.get('/metrics', async (req: Request, res: Response) => {
  res.set('Content-Type', prometheusService.getContentType());
  res.end(await prometheusService.getMetrics());
});

// Bull Board dashboard (development only)
if (config.app.environment === 'development') {
  app.use('/admin/queues', bullAdapter);
}

// Health check endpoints
app.get('/health', healthRoutes.getHealth);
app.get('/health/detailed', healthRoutes.getDetailedHealth);
app.get('/ready', healthRoutes.getReady);
app.get('/live', healthRoutes.getLive);

// API routes
app.use('/api/v1/reports', reportRoutes);
app.use('/api/v1/schedules', scheduleRoutes);
app.use('/api/v1/templates', templateRoutes);

// Error handling middleware
app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  console.error('Error:', err);
  
  prometheusService.incrementErrorCounter('express');
  
  res.status(500).json({
    status: 'error',
    message: config.app.environment === 'development' ? err.message : 'Internal server error',
    code: 'INTERNAL_ERROR',
  });
});

// 404 handler
app.use((req: Request, res: Response) => {
  res.status(404).json({
    status: 'error',
    message: 'Not found',
    code: 'NOT_FOUND',
  });
});

// Start server
async function startServer(): Promise<void> {
  try {
    // Initialize Kafka
    await kafkaService.connect();
    console.log('Kafka connected');

    // Start report queue consumer
    await reportQueue.startConsumer();
    console.log('Report queue consumer started');

    // Start Express server
    const port = config.server.port;
    app.listen(port, () => {
      console.log(`🚀 Regulatory Reporting Service running on port ${port}`);
      console.log(`📚 API Documentation: http://localhost:${port}/api/v1/docs`);
      console.log(`📊 Bull Board Dashboard: http://localhost:${port}/admin/queues`);
      console.log(`🌍 Environment: ${config.app.environment}`);
    });

  } catch (error) {
    console.error('Failed to start server:', error);
    process.exit(1);
  }
}

// Graceful shutdown
async function shutdown(signal: string): Promise<void> {
  console.log(`\n${signal} received, shutting down gracefully...`);

  try {
    await reportQueue.close();
    await kafkaService.disconnect();
    console.log('Shutdown complete');
    process.exit(0);
  } catch (error) {
    console.error('Error during shutdown:', error);
    process.exit(1);
  }
}

process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT', () => shutdown('SIGINT'));

// Start the server
startServer();

export { app };
