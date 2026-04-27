package accrual

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/aikowocki/yandex-go-first-diploma/internal/entity"
	"go.uber.org/zap"
)

type AccrualProcessor interface {
	GetPendingOrders(ctx context.Context) ([]entity.Order, error)
	ProcessOrder(ctx context.Context, order entity.Order) error
}

type WorkerPool struct {
	processor  AccrualProcessor
	numWorkers int
}

func NewWorkerPool(processor AccrualProcessor, numWorkes int) *WorkerPool {
	return &WorkerPool{
		processor:  processor,
		numWorkers: numWorkes,
	}
}

func (w *WorkerPool) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processOrders(ctx)
		}
	}

}

func (w *WorkerPool) processOrders(ctx context.Context) {
	orders, err := w.processor.GetPendingOrders(ctx)
	if err != nil {
		zap.S().Errorw("failed to find pending orders", "error", err)
		return
	}
	zap.S().Infow("found pending orders", "count", len(orders))

	if len(orders) == 0 {
		return
	}

	jobs := make(chan entity.Order, len(orders))
	for _, order := range orders {
		jobs <- order
	}
	close(jobs)

	pauseCh := make(chan time.Duration, 1)

	var wg sync.WaitGroup
	for i := 0; i < w.numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.worker(ctx, jobs, pauseCh)
		}()
	}

	wg.Wait()
}

func (w *WorkerPool) worker(ctx context.Context, jobs <-chan entity.Order, pauseCh chan time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		case duration := <-pauseCh:
			// 429
			select {
			case pauseCh <- duration:
			default:
			}
			select {
			case <-time.After(duration):
			case <-ctx.Done():
				return
			}
		case order, ok := <-jobs:
			if !ok {
				return
			}
			if err := w.processor.ProcessOrder(ctx, order); err != nil {
				var tooMany *ErrTooManyRequests
				if errors.As(err, &tooMany) {
					select {
					case pauseCh <- tooMany.RetryAfter:
					default:
					}
					continue
				}
				zap.S().Errorw("failed to process order", "error", err, "order", order.Number)
			}
		}
	}
}
