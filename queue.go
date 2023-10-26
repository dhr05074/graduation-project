package main

import "sync"

type Request struct {
	ch       *chan struct{}
	priority int
}

type PriorityQueue []*Request

func (pq *PriorityQueue) Len() int {
	return len(*pq)
}

func (pq *PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the request with the lowest priority,
	// so we use less than here.
	return (*pq)[i].priority < (*pq)[j].priority
}

func (pq *PriorityQueue) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Request)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

type Queue struct {
	pq *PriorityQueue
	sync.Mutex
}
