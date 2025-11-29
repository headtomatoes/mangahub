package mangadex

import (
	"context"
	"log"
	"sync"
)

// Task represents a unit of work to be processed by the worker pool
type Task func(ctx context.Context) error

// WorkerPool manages concurrent processing of tasks
type WorkerPool struct {
	workerCount int
	taskQueue   chan Task
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	closed      bool
	closeMux    sync.Mutex
}

// NewWorkerPool creates a pool with specified number of workers
func NewWorkerPool(workerCount int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan Task, workerCount*2), // Buffered channel
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start launches worker goroutines
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	log.Printf("[WorkerPool] Started %d workers", wp.workerCount)
}

// Submit adds a task to the queue (non-blocking with context check)
func (wp *WorkerPool) Submit(task Task) {
	select {
	case wp.taskQueue <- task:
		// Task submitted successfully
	case <-wp.ctx.Done():
		// Pool is shutting down
		log.Println("[WorkerPool] Pool shutting down, task not submitted")
	}
}

// Wait blocks until all tasks complete
func (wp *WorkerPool) Wait() {
	wp.closeMux.Lock()
	if !wp.closed {
		close(wp.taskQueue) // No more tasks
		wp.closed = true
	}
	wp.closeMux.Unlock()

	wp.wg.Wait() // Wait for workers to finish
	log.Println("[WorkerPool] All workers completed")
}

// Shutdown cancels all workers and waits for completion
func (wp *WorkerPool) Shutdown() {
	log.Println("[WorkerPool] Shutting down...")
	wp.cancel()
	wp.Wait()
}

// worker processes tasks from the queue
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for task := range wp.taskQueue {
		// Check if context is cancelled
		select {
		case <-wp.ctx.Done():
			log.Printf("[Worker %d] Context cancelled, stopping", id)
			return
		default:
		}

		// Execute task
		if err := task(wp.ctx); err != nil {
			log.Printf("[Worker %d] Task error: %v", id, err)
		}
	}
}

// WorkerPoolWithContext creates a worker pool with a custom context
func WorkerPoolWithContext(ctx context.Context, workerCount int) *WorkerPool {
	poolCtx, cancel := context.WithCancel(ctx)
	return &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan Task, workerCount*2),
		ctx:         poolCtx,
		cancel:      cancel,
	}
}
