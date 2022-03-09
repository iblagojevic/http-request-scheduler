package requestscheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var SFQ *ScheduledFunctionsQueue

func HandleIncomingRequests(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var msg Message
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// enqueue message
	now := time.Now()
	executionTime := now.Add(time.Second * time.Duration(msg.Delay))
	sendMessageToQueue(executionTime, msg.Action, msg.Url, msg.Payload)
	fmt.Fprintf(w, "accepted")
}

func InitQueue() *ScheduledFunctionsQueue {
	ctx := context.Background()
	q := NewScheduledFunctionsQueue(ctx)
	q.StartProcessingScheduledFunctions()
	return q
}

func sendMessageToQueue(executionTime time.Time, args ...interface{}) {
	SFQ.Push(executionTime, Executable, args...)
}
