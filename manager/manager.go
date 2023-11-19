package manager

import (
	"errors"
	"sync"
	"time"
)

type BucketManager struct {
	bucket       uint
	bucketSize   int
	fillInterval time.Duration

	mu sync.Mutex

	killed bool
}

func NewBucketManager(bucketSize int, fillInterval time.Duration) *BucketManager {
	return &BucketManager{
		bucketSize:   bucketSize,
		fillInterval: fillInterval,
		killed:       false,
		mu:           sync.Mutex{},
	}
}

func (m *BucketManager) Start() {
	go m.fillBucket()
}

func (m *BucketManager) fillBucket() {
	for !m.killed {
		m.mu.Lock()
		m.bucket += 1
		m.mu.Unlock()

		time.Sleep(m.fillInterval)
	}
}

func (m *BucketManager) Consume(priority uint) error {
	if m.killed {
		return errors.New("bucket manager is killed")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.bucket < priority {
		return errors.New("not enough bucket")
	}

	m.bucket -= priority

	return nil
}

func (m *BucketManager) Kill() {
	m.killed = true
}
