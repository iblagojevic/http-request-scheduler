package requestscheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

/*
This module defines model into which request is converted (Message) and function that implements ScheduledFunction signature
For different purposes (where scheduler is not used to make delayed http requests), this module should be changed accordingly
*/

type Message struct {
	Action  string                 `json:"action"`
	Url     string                 `json:"url"`
	Payload map[string]interface{} `json:"payload"`
	Delay   int                    `json:"delay"`
}

var Executable = func(args ...interface{}) {
	// Specific for this executable. For your own do proper casting.
	if len(args) > 3 {
		fmt.Println("Executable expects 3 arguments")
	}
	var msg Message
	for i, arg := range args {
		switch i {
		case 0: // action
			action, ok := arg.(string)
			if !ok {
				fmt.Println("action is not string")
			}
			msg.Action = action
		case 1:
			url, ok := arg.(string)
			if !ok {
				fmt.Println("url is not string")
			}
			msg.Url = url
		case 2:
			payload, ok := arg.(map[string]interface{})
			if !ok {
				fmt.Println("payload is not string")
			}
			msg.Payload = payload
		default:
			fmt.Println("Executable expects 3 arguments")
		}
	}
	switch strings.ToLower(msg.Action) {
	case "post":
		payload, err1 := json.Marshal(msg.Payload)
		if err1 != nil {
			fmt.Printf("Cannot process sent payload %s for url %s. %v\n", payload, msg.Url, err1)
		}
		requestBody := bytes.NewBuffer(payload)
		_, err2 := http.Post(msg.Url, "application/json", requestBody)
		if err2 != nil {
			fmt.Printf("An error occured POSTing to %s. %v\n", msg.Url, err2)
		}
	case "get":
		_, err := http.Get(msg.Url)
		if err != nil {
			fmt.Printf("An error occured GETing from %s. %v\n", msg.Url, err)
		}
	default:
		fmt.Printf("Unknown action %s\n", msg.Action)
	}
}
