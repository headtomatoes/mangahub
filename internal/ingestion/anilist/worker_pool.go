package anilist

import (
    "context"
    "log"
    "sync"
)

// Task represents a unit of work
type Task func(ctx context.Context) error

// WorkerPool manages concurrent processing
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
        taskQueue:   make(chan Task, workerCount*2),
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
    log.Printf("[AniListWorkerPool] Started %d workers", wp.workerCount)
}

// Submit adds a task to the queue
func (wp *WorkerPool) Submit(task Task) {
    select {
    case wp.taskQueue <- task:
    case <-wp.ctx.Done():
        log.Println("[AniListWorkerPool] Pool is shutting down, task rejected")
    }
}

// Wait blocks until all tasks complete
func (wp *WorkerPool) Wait() {
    wp.closeMux.Lock()
    if !wp.closed {
        close(wp.taskQueue)
        wp.closed = true
    }
    wp.closeMux.Unlock()

    wp.wg.Wait()
    log.Println("[AniListWorkerPool] All workers completed")
}

// Shutdown cancels all workers
func (wp *WorkerPool) Shutdown() {
    log.Println("[AniListWorkerPool] Shutting down...")
    wp.cancel()
    wp.Wait()
}

// worker processes tasks from the queue
func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()

    for {
        select {
        case task, ok := <-wp.taskQueue:
            if !ok {
                return
            }

            if err := task(wp.ctx); err != nil {
                log.Printf("[AniListWorker-%d] Task failed: %v", id, err)
            }

        case <-wp.ctx.Done():
            return
        }
    }
}