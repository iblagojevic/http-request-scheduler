package requestscheduler

import (
	"container/heap"
	"context"
	"time"
)

/*
Module defines structures and operations on ScheduledFunctionsQueue

Structures:

EnqueuedMessage - container for scheduled function with its arguments and time of its execution
ScheduledFunction - prototype for functions to be executed
ScheduledFunctionsQueue - container for PriorityQueue and control structures needed to manage PriorityQueue

Functions:

NewScheduledFunctionsQueue - instantiates ScheduledFunctionsQueue and initializes internal PriorityQueue
StartProcessingScheduledFunctions - starts goroutine to process enqueued messages from PriorityQueue; each scheduled function will be handled by separate goroutine
Push - wrapper for adding new scheduled functions to queue via channel
enqueue - adds new message with scheduled function and time of execution to the PriorityQueue
dequeue - scans PriorityQueue for scheduled functions eligible for execution
*/

// EnqueuedMessage contains function executed at given time by the TimerQueue.
type EnqueuedMessage struct {
	fn            ScheduledFunction
	args          []interface{}
	executionTime time.Time
}

// ScheduledFunction is the function invoked by EnqueuedMessage struct when time comes
type ScheduledFunction func(args ...interface{})

// ScheduledFunctionsQueue is priority queue of scheduled functions
type ScheduledFunctionsQueue struct {
	pq          PriorityQueue
	ctx         context.Context
	nextRunAt   *time.Time
	pushChannel chan *EnqueuedMessage
	timer       *time.Timer
}

// NewScheduledFunctionsQueue creates a new instance of ScheduledFunctionsQueue
func NewScheduledFunctionsQueue(ctx context.Context) *ScheduledFunctionsQueue {
	q := &ScheduledFunctionsQueue{
		pq:          PriorityQueue{},
		ctx:         ctx,
		pushChannel: make(chan *EnqueuedMessage, 8192),
		timer:       time.NewTimer(time.Second * 86400),
	}
	// initialize priority queue which is element of schedulesd functions queue
	heap.Init(&q.pq)
	return q
}

// Push adds to-be-executed function to the queue. Returns true or false (if context is cancelled)
func (q *ScheduledFunctionsQueue) Push(executionTime time.Time, fn ScheduledFunction, args ...interface{}) bool {
	select {
	case q.pushChannel <- &EnqueuedMessage{executionTime: executionTime, fn: fn, args: args}:
		return true
	case <-q.ctx.Done():
		return false
	}
}

func (q *ScheduledFunctionsQueue) Drain() {
	// runs in main thread where shutdown is handled
	now := time.Now()
	executables := q.pop(now, true)
	for _, toExecute := range executables {
		toExecute.fn(toExecute.args...)
	}
}

// StartProcessingScheduledFunctions start goroutine within the given context that processes enqueued messages. New goroutine executes for each message.
func (q *ScheduledFunctionsQueue) StartProcessingScheduledFunctions() {
	go func() {
		for {
			select {
			case <-q.ctx.Done():
				return
			case now := <-q.timer.C:
				executables := q.pop(now, false)
				// run the functions in a different goroutine so this one remains free for accepting push calls
				go func() {
					for _, toExecute := range executables {
						toExecute.fn(toExecute.args...)
					}
				}()

				if q.pq.Len() > 0 {
					nextRun := q.pq[0].executionTime
					q.nextRunAt = &nextRun
					q.timer.Reset(time.Until(nextRun))
				} else {
					q.timer.Stop()
					q.nextRunAt = nil
				}
			case msg := <-q.pushChannel:
				q.push(msg)
			}
		}
	}()
}

func (q *ScheduledFunctionsQueue) push(msg *EnqueuedMessage) {
	heap.Push(&q.pq, msg)
	if q.nextRunAt == nil || q.nextRunAt.After(msg.executionTime) {
		if q.nextRunAt != nil && !q.timer.Stop() {
			<-q.timer.C
		}
		q.timer.Reset(time.Until(msg.executionTime))
		q.nextRunAt = &msg.executionTime
	}
}

func (q *ScheduledFunctionsQueue) pop(now time.Time, all bool) (result []*EnqueuedMessage) {
	for i := 0; q.pq.Len() > 0; i++ {
		current := q.pq[0]
		// take all messages whose execution time passed or literally all if "all" flag is set
		if all || current.executionTime.Before(now) {
			msg := heap.Pop(&q.pq).(*EnqueuedMessage)
			result = append(result, msg)
		} else {
			break
		}
	}
	return result
}
