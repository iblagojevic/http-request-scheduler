package requestscheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

type PayloadForTarget struct {
	Object string `json:"object"`
	Id     int    `json:"id"`
}

var targetResultIds []int
var targetResultObjects []string

func TargetEndpointPOST(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	var msg PayloadForTarget
	json.Unmarshal(b, &msg)
	targetResultIds = append(targetResultIds, msg.Id)
	targetResultObjects = append(targetResultObjects, msg.Object)
	fmt.Fprintf(w, "ok")
}

func TargetEndpointGET(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	targetResultIds = append(targetResultIds, id)
	targetResultObjects = append(targetResultObjects, r.URL.Query().Get("object"))
	fmt.Fprintf(w, "ok")
}

func serverAlive(port string) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", port), timeout)
	if err != nil {
		return false
	}
	if conn != nil {
		defer conn.Close()
		return true
	}
	return false
}

func TestEndToEnd(t *testing.T) {
	var serverPort = "9293"
	var targetPort = "9393"
	SFQ = InitQueue()
	router := mux.NewRouter()
	targetRouter := mux.NewRouter()
	router.HandleFunc("/", HandleIncomingRequests).Methods("POST")
	targetRouter.HandleFunc("/post", TargetEndpointPOST).Methods("POST")
	targetRouter.HandleFunc("/get", TargetEndpointGET).Methods("GET")

	// init listener
	webserver := &http.Server{
		Addr:    ":" + serverPort,
		Handler: router,
	}

	// init target webserver
	targetWebserver := &http.Server{
		Addr:    ":" + targetPort,
		Handler: targetRouter,
	}
	go func() {
		if err := webserver.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error in webserver: %s\n", err)
		}
	}()

	go func() {
		if err := targetWebserver.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error in target webserver: %s\n", err)
		}
	}()

	for {
		time.Sleep(100 * time.Millisecond)
		if serverAlive(serverPort) {
			break
		}
	}

	for {
		time.Sleep(100 * time.Millisecond)
		if serverAlive(targetPort) {
			break
		}
	}

	// fire POST requests to listener
	for i := 1; i <= 10; i++ {
		pft := make(map[string]interface{})
		pft["id"] = i
		pft["object"] = "Event_" + strconv.Itoa(i)
		pfl := Message{Action: "POST", Url: "http://localhost:" + targetPort + "/post", Payload: pft, Delay: rand.Intn(5)}
		requestBody, _ := json.Marshal(pfl)
		http.Post("http://localhost:"+serverPort+"/", "application/json", bytes.NewBuffer(requestBody))
	}

	// wait some time before all delayed messages are processed
	time.Sleep(10 * time.Second)

	// assert message ended up on target server
	require.Equal(t, 10, len(targetResultIds))
	require.Equal(t, 10, len(targetResultObjects))
	var found = 0

	for _, a := range targetResultObjects {
		for j := 1; j <= 10; j++ {
			if a == "Event_"+strconv.Itoa(j) {
				found++
			}
		}
	}
	require.Equal(t, 10, found)

	// close both listener and target
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	webserver.Shutdown(ctx)
	targetWebserver.Shutdown(ctx)
}
