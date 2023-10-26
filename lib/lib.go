package lib

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type QueueManager struct {
	inboundQueue chan int
	levelQueue   map[int]chan struct{}
	processQueue chan int

	mu *sync.Mutex

	priorities []int

	workers int

	throughput map[int]int

	nowProcessedCnt int
	nowIncomeCnt    int

	kill chan struct{}

	increment       chan int
	incrementIncome chan int
	clearCnt        chan struct{}
}

func NewQueueManager(processQueue chan int) *QueueManager {
	return &QueueManager{
		processQueue:    processQueue,
		levelQueue:      make(map[int]chan struct{}),
		workers:         20,
		inboundQueue:    make(chan int, 1<<20),
		throughput:      make(map[int]int),
		priorities:      []int{},
		mu:              &sync.Mutex{},
		kill:            make(chan struct{}),
		increment:       make(chan int, 1<<10),
		incrementIncome: make(chan int, 1<<10),
		clearCnt:        make(chan struct{}, 1<<10),
	}
}

func (m *QueueManager) Wait() {
	<-m.kill
}

func (m *QueueManager) Start() chan int {
	for i := 0; i < m.workers; i++ {
		go func() {
			for {
				p, ok := <-m.processQueue
				if !ok {
					return
				}
				m.increment <- p
				time.Sleep(10 * time.Nanosecond)
			}
		}()
	}

	go func() {
		tick := time.Tick(1 * time.Second)
		for {
			if m.inboundQueue == nil {
				return
			}

			select {
			case p := <-m.increment:
				m.throughput[p] += 1
				m.nowProcessedCnt += 1
			case <-m.incrementIncome:
				m.nowIncomeCnt += 1
			case <-m.clearCnt:
				m.throughput = map[int]int{}
				m.nowIncomeCnt = 0
				m.nowProcessedCnt = 0
			case <-tick:
				stat := fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%.2f\n", m.throughput[1], m.throughput[2], m.throughput[3], m.throughput[4], m.throughput[5], m.nowProcessedCnt, m.nowIncomeCnt, float64(m.nowProcessedCnt)/float64(m.nowIncomeCnt))
				fmt.Printf(stat)
			}
		}
	}()

	go func() {
		for {
			select {
			case p, ok := <-m.inboundQueue:
				if !ok {
					return
				}

				m.incrementIncome <- p

				q, ok := m.levelQueue[p]
				if !ok {
					q = make(chan struct{}, 1<<(10-p))
					m.levelQueue[p] = q

					m.priorities = append(m.priorities, p)
					sort.Ints(m.priorities)
				}

				select {
				case q <- struct{}{}:
				default:
				}
			default:
				for _, p := range m.priorities {
					select {
					case _, ok := <-m.levelQueue[p]:
						if !ok {
							return
						}
						m.processQueue <- p
					default:
					}
				}
			}
		}
	}()

	return m.inboundQueue
}
