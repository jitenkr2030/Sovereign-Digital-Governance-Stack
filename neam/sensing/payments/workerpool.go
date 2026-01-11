package main

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
	workerCount int
	jobs        chan func()
	logger      *zap.Logger
	wg          sync.WaitGroup
	shutdown    chan struct{}
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workerCount int, logger *zap.Logger) *WorkerPool {
	pool := &WorkerPool{
		workerCount: workerCount,
		jobs:        make(chan func(), workerCount*10),
		logger:      logger,
		shutdown:    make(chan struct{}),
	}

	return pool
}

// Start starts the worker pool
func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}

	p.logger.Info("Worker pool started",
		zap.Int("worker_count", p.workerCount),
		zap.Int("queue_size", p.workerCount*10))
}

// Submit submits a job to the worker pool
func (p *WorkerPool) Submit(job func()) error {
	select {
	case p.jobs <- job:
		return nil
	default:
		return ErrPoolFull
	}
}

// SubmitWithTimeout submits a job with timeout
func (p *WorkerPool) SubmitWithTimeout(job func(), timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case p.jobs <- job:
		return nil
	case <-ctx.Done():
		return ErrPoolTimeout
	}
}

// Stop gracefully stops the worker pool
func (p *WorkerPool) Stop() error {
	close(p.shutdown)
	p.wg.Wait()
	close(p.jobs)

	p.logger.Info("Worker pool stopped")
	return nil
}

// worker is the main worker goroutine
func (p *WorkerPool) worker(ctx context.Context, id int) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			p.logger.Debug("Worker context cancelled",
				zap.Int("worker_id", id))
			return

		case <-p.shutdown:
			p.logger.Debug("Worker shutdown signal received",
				zap.Int("worker_id", id))
			return

		case job, ok := <-p.jobs:
			if !ok {
				return
			}

			// Execute job with panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						p.logger.Error("Worker panic recovered",
							zap.Int("worker_id", id),
							zap.Any("panic", r))
					}
				}()

				start := time.Now()
				job()
				duration := time.Since(start)

				// Log slow jobs
				if duration > 1*time.Second {
					p.logger.Warn("Slow job completed",
						zap.Int("worker_id", id),
						zap.Duration("duration", duration))
				}
			}()
		}
	}
}

// Stats returns current worker pool statistics
func (p *WorkerPool) Stats() WorkerStats {
	return WorkerStats{
		WorkerCount: p.workerCount,
		QueueLength: len(p.jobs),
	}
}

// WorkerStats contains worker pool statistics
type WorkerStats struct {
	WorkerCount int
	QueueLength int
}

// Errors
var (
	ErrPoolFull    = fmt.Errorf("worker pool queue is full")
	ErrPoolTimeout = fmt.Errorf("worker pool submission timed out")
)
