package ohlcv

import (
	"context"
	"sync"
	"sync/atomic"
)

type WorkerPool struct {
	workers   int
	taskCh    chan func()
	stopCh    chan struct{}
	wg        sync.WaitGroup
	isRunning int32
}

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		taskCh:  make(chan func(), workers*2),
		stopCh:  make(chan struct{}),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&wp.isRunning, 0, 1) {
		return
	}

	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx)
	}
}

func (wp *WorkerPool) worker(ctx context.Context) {
	defer wp.wg.Done()

	for {
		select {
		case task := <-wp.taskCh:
			if task != nil {
				task()
			}
		case <-ctx.Done():
			return
		case <-wp.stopCh:
			return
		}
	}
}

func (wp *WorkerPool) Submit(task func()) bool {
	if atomic.LoadInt32(&wp.isRunning) == 0 {
		return false
	}

	select {
	case wp.taskCh <- task:
		return true
	default:
		return false
	}
}

func (wp *WorkerPool) Stop() {
	if !atomic.CompareAndSwapInt32(&wp.isRunning, 1, 0) {
		return
	}

	close(wp.stopCh)
	wp.wg.Wait()
}
