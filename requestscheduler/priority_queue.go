package requestscheduler

/**
Module defines PriorityQueue and relevant operations
**/

type PriorityQueue []*EnqueuedMessage

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].executionTime.Before(pq[j].executionTime)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(m interface{}) {
	msg := m.(*EnqueuedMessage)
	*pq = append(*pq, msg)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	msg := old[n-1]
	old[n-1] = nil
	*pq = old[0 : n-1]
	return msg
}
