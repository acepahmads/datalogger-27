package queue

import (
	"errors"
	"sync"
)

var (
	ErrQueueFull  = errors.New("queue is at maximum capacity")
	ErrQueueEmpty = errors.New("queue is empty")
)

type LocalQueue struct {
	mu          sync.Mutex
	items       []interface{}
	capacity    int
	totalPushed int64
	totalPopped int64
	totalDrop   int64
}

func NewLocalQueue(capacity int) *LocalQueue {
	if capacity <= 0 {
		capacity = 10000
	}
	return &LocalQueue{
		items:    make([]interface{}, 0, 128),
		capacity: capacity,
	}
}

func (q *LocalQueue) Enqueue(item interface{}) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= q.capacity {
		q.totalDrop++
		return ErrQueueFull
	}

	q.items = append(q.items, item)
	q.totalPushed++
	return nil
}

func (q *LocalQueue) Dequeue() (interface{}, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil, ErrQueueEmpty
	}

	item := q.items[0]
	q.items = q.items[1:]
	q.totalPopped++
	return item, nil
}

func (q *LocalQueue) DequeueBatch(maxBatch int) []interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	count := maxBatch
	if count > len(q.items) {
		count = len(q.items)
	}

	batch := make([]interface{}, count)
	copy(batch, q.items[:count])
	q.items = q.items[count:]
	q.totalPopped += int64(count)
	return batch
}

func (q *LocalQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

func (q *LocalQueue) Stats() map[string]interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	return map[string]interface{}{
		"current_len":  len(q.items),
		"capacity":     q.capacity,
		"total_pushed": q.totalPushed,
		"total_popped": q.totalPopped,
		"total_drops":  q.totalDrop,
	}
}
